package tasks

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"encoding/xml"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NmapTask struct {
	TaskTemplate
	resultDAO *dao.ResultDAO
}

func NewNmapTask(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *NmapTask {
	return &NmapTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
		resultDAO:    resultDAO,
	}
}

func (n *NmapTask) Handle(ctx context.Context, t *asynq.Task) error {
	return n.Execute(ctx, t, n.runNmap)
}

func (n *NmapTask) runNmap(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Host     string `json:"target"`
		TaskID   string `json:"task_id"`
		ParentID string `json:"parent_id,omitempty"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Host == "" {
		return fmt.Errorf("无效的主机地址")
	}

	logging.Info("开始执行 Nmap 任务: %s", payload.Host)

	// 创建临时文件来存储 Nmap 结果
	tempFile, err := os.CreateTemp("", "nmap-result-*.xml")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 构建 Nmap 命令
	cmd := exec.CommandContext(ctx, "nmap",
		"-n", "--unique", "--resolve-all", "-Pn",
		"--min-hostgroup", "64",
		"--max-retries", "0",
		"--host-timeout", "10m",
		"--script-timeout", "3m",
		"-oX", tempFile.Name(),
		"--version-intensity", "9",
		"--min-rate", "10000",
		"-T4",
		payload.Host,
	)

	// 执行 Nmap 命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行 Nmap 命令失败: %v, 输出: %s", err, string(output))
	}

	// 读取 XML 结果
	xmlData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("读取 Nmap 结果文件失败: %v", err)
	}

	// 解析 XML 结果
	portEntries, err := parseNmapXML(xmlData)
	if err != nil {
		return fmt.Errorf("解析 Nmap XML 结果失败: %v", err)
	}

	portData := &models.PortData{
		Ports: portEntries,
	}

	var parentID *primitive.ObjectID
	if payload.ParentID != "" {
		objID, err := primitive.ObjectIDFromHex(payload.ParentID)
		if err == nil {
			parentID = &objID
		}
	}

	scanResult := &models.Result{
		ID:        primitive.NewObjectID(),
		Type:      "Port",
		Target:    payload.Host,
		Timestamp: time.Now(),
		Data:      portData,
		ParentID:  parentID,
	}

	if err := n.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	logging.Info("Nmap 任务完成，扫描了 %d 个端口", len(portEntries))

	return nil
}

func parseNmapXML(data []byte) ([]models.PortEntry, error) {
	var result struct {
		Hosts []struct {
			Addresses []struct {
				Addr     string `xml:"addr,attr"`
				AddrType string `xml:"addrtype,attr"`
			} `xml:"address"`
			Ports []struct {
				ID       int    `xml:"portid,attr"`
				Protocol string `xml:"protocol,attr"`
				State    struct {
					State string `xml:"state,attr"`
				} `xml:"state"`
				Service struct {
					Name      string `xml:"name,attr"`
					Product   string `xml:"product,attr"`
					Version   string `xml:"version,attr"`
					ExtraInfo string `xml:"extrainfo,attr"`
				} `xml:"service"`
			} `xml:"ports>port"`
		} `xml:"host"`
	}

	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var portEntries []models.PortEntry
	for _, host := range result.Hosts {
		// 获取主机 IP 地址
		var hostAddr string
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" {
				hostAddr = addr.Addr
				break
			}
		}

		// 如果没有找到 IPv4 地址，尝试使用任意类型的地址
		if hostAddr == "" && len(host.Addresses) > 0 {
			hostAddr = host.Addresses[0].Addr
		}

		for _, port := range host.Ports {
			portEntries = append(portEntries, models.PortEntry{
				ID:         primitive.NewObjectID(),
				Host:       hostAddr, // 添加主机地址
				Number:     port.ID,
				Protocol:   port.Protocol,
				State:      port.State.State,
				Service:    port.Service.Name,
				IsRead:     false,
				HTTPStatus: 0,
				HTTPTitle:  "",
			})
		}
	}

	return portEntries, nil
}
