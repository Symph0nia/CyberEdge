// dns_service.go

package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/utils"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

type DNSService struct {
	resultDAO *dao.ResultDAO
}

func NewDNSService(resultDAO *dao.ResultDAO) *DNSService {
	return &DNSService{
		resultDAO: resultDAO,
	}
}

func (s *DNSService) ResolveAndUpdateSubdomainIP(resultID, entryID string) error {
	// 获取扫描结果
	result, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return err
	}

	// 检查任务类型
	if result.Type != "Subdomain" {
		return errors.New("任务类型不是子域名扫描")
	}

	// 解析 SubdomainData
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(result.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return err
	}

	// 查找指定的子域名
	var targetSubdomain *models.SubdomainEntry
	for i := range subdomainData.Subdomains {
		if subdomainData.Subdomains[i].ID.Hex() == entryID {
			targetSubdomain = &subdomainData.Subdomains[i]
			break
		}
	}

	if targetSubdomain == nil {
		return errors.New("未找到指定的子域名")
	}

	resolvedIP, err := s.resolveIPWithFallback(targetSubdomain.Domain)
	if err != nil {
		logging.Error("解析域名失败: %v", err)
		return err
	}

	if resolvedIP != "" {
		err = s.resultDAO.UpdateSubdomainIP(resultID, entryID, resolvedIP)
		if err != nil {
			logging.Error("更新子域名 IP 失败: %v", err)
			return err
		}
		logging.Info("成功更新子域名 %s 的 IP 为 %s", targetSubdomain.Domain, resolvedIP)
	} else {
		logging.Warn("未能解析到 IP 地址: %s", targetSubdomain.Domain)
	}

	return nil
}

func (s *DNSService) resolveIPWithFallback(domain string) (string, error) {
	// 定义 DNS 服务器列表
	dnsServers := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}

	for _, server := range dnsServers {
		ip, err := s.resolveIPUsingDNSServer(domain, server)
		if err == nil && ip != "" {
			return ip, nil
		}
		logging.Warn("使用 DNS 服务器 %s 解析域名 %s 失败: %v", server, domain, err)
	}

	// 如果所有 DNS 服务器都失败，尝试使用系统默认 DNS
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}
	if len(ips) > 0 {
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}

	return "", errors.New("无法解析域名")
}

func (s *DNSService) resolveIPUsingDNSServer(domain, dnsServer string) (string, error) {
	c := dns.Client{Timeout: 2 * time.Second}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := c.Exchange(&m, dnsServer+":53")
	if err != nil {
		return "", err
	}

	if len(r.Answer) == 0 {
		return "", errors.New("no answer")
	}

	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			return a.A.String(), nil
		}
	}

	return "", errors.New("no A record found")
}

type ResolveResult struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"` // entryId -> error message
}

func (s *DNSService) BatchResolveAndUpdateSubdomainIP(resultID string, entryIDs []string) (*ResolveResult, error) {
	logging.Info("开始批量解析子域名IP: %v", entryIDs)

	result := &ResolveResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 获取扫描结果
	scanResult, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return nil, err
	}

	// 检查任务类型
	if scanResult.Type != "Subdomain" {
		return nil, errors.New("任务类型不是子域名扫描")
	}

	// 解析 SubdomainData
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(scanResult.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return nil, err
	}

	// 创建entryID到subdomain的映射
	subdomainMap := make(map[string]*models.SubdomainEntry)
	for i := range subdomainData.Subdomains {
		subdomainMap[subdomainData.Subdomains[i].ID.Hex()] = &subdomainData.Subdomains[i]
	}

	// 批量解析和更新
	for _, entryID := range entryIDs {
		subdomain, exists := subdomainMap[entryID]
		if !exists {
			result.Failed[entryID] = "未找到指定的子域名"
			continue
		}

		// 如果已经有IP，跳过
		if subdomain.IP != "" {
			continue
		}

		resolvedIP, err := s.resolveIPWithFallback(subdomain.Domain)
		if err != nil {
			result.Failed[entryID] = fmt.Sprintf("解析失败: %v", err)
			continue
		}

		if resolvedIP == "" {
			result.Failed[entryID] = "未能解析到IP地址"
			continue
		}

		// 更新IP
		err = s.resultDAO.UpdateSubdomainIP(resultID, entryID, resolvedIP)
		if err != nil {
			result.Failed[entryID] = fmt.Sprintf("更新IP失败: %v", err)
			continue
		}

		result.Success = append(result.Success, entryID)
		logging.Info("成功更新子域名 %s 的 IP 为 %s", subdomain.Domain, resolvedIP)
	}

	logging.Info("批量解析完成，成功: %d, 失败: %d",
		len(result.Success), len(result.Failed))

	return result, nil
}

func (s *DNSService) ResolveSubdomainIPs(resultID string, entryIDs []string) (*ResolveResult, error) {
	logging.Info("开始解析子域名IP: %v", entryIDs)

	result := &ResolveResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 获取扫描结果和数据验证
	scanResult, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return nil, err
	}

	if scanResult.Type != "Subdomain" {
		return nil, errors.New("任务类型不是子域名扫描")
	}

	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(scanResult.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return nil, err
	}

	// 创建 ID 映射
	subdomainMap := make(map[string]*models.SubdomainEntry)
	for i := range subdomainData.Subdomains {
		subdomainMap[subdomainData.Subdomains[i].ID.Hex()] = &subdomainData.Subdomains[i]
	}

	// 创建并发控制
	type resolveResult struct {
		EntryID string
		IP      string
		Error   error
	}

	maxWorkers := 10 // 最大并发数
	resultChan := make(chan resolveResult, len(entryIDs))
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// 启动解析协程
	for _, entryID := range entryIDs {
		subdomain, exists := subdomainMap[entryID]
		if !exists {
			result.Failed[entryID] = "未找到指定的子域名"
			continue
		}

		// 跳过已有IP的记录
		if subdomain.IP != "" {
			continue
		}

		wg.Add(1)
		go func(entryID string, domain string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 解析IP
			resolvedIP, err := s.resolveIPWithFallback(domain)
			resultChan <- resolveResult{
				EntryID: entryID,
				IP:      resolvedIP,
				Error:   err,
			}
		}(entryID, subdomain.Domain)
	}

	// 启动结果处理协程
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 处理结果
	var mu sync.Mutex // 用于保护 result 的并发访问
	for res := range resultChan {
		if res.Error != nil {
			mu.Lock()
			result.Failed[res.EntryID] = fmt.Sprintf("解析失败: %v", res.Error)
			mu.Unlock()
			continue
		}

		if res.IP == "" {
			mu.Lock()
			result.Failed[res.EntryID] = "未能解析到IP地址"
			mu.Unlock()
			continue
		}

		// 更新IP
		if err := s.resultDAO.UpdateSubdomainIP(resultID, res.EntryID, res.IP); err != nil {
			mu.Lock()
			result.Failed[res.EntryID] = fmt.Sprintf("更新IP失败: %v", err)
			mu.Unlock()
			continue
		}

		mu.Lock()
		result.Success = append(result.Success, res.EntryID)
		mu.Unlock()

		logging.Info("成功更新子域名IP: %s -> %s", subdomainMap[res.EntryID].Domain, res.IP)
	}

	logging.Info("解析完成，成功: %d, 失败: %d", len(result.Success), len(result.Failed))
	return result, nil
}
