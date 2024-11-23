// dns_service.go

package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/utils"
	"errors"
	"github.com/miekg/dns"
	"net"
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
