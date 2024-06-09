from django.db import models
from common.models import ScanJob
from django.contrib.contenttypes.models import ContentType

class Subdomain(models.Model):
    scan_job = models.ForeignKey(ScanJob, on_delete=models.CASCADE, related_name='subdomains',
                                 verbose_name='子域名扫描任务')
    subdomain = models.CharField(max_length=255, verbose_name='子域名')
    domain = models.CharField(max_length=255, verbose_name='域名')
    source = models.CharField(max_length=255, verbose_name='来源', blank=True)
    ip_address = models.CharField(max_length=100, verbose_name='IP地址', null=True, blank=True)

    # HTTP和HTTPS状态码
    subdomain_http_status = models.CharField(max_length=10, verbose_name='子域名HTTP状态码', null=True, blank=True)
    subdomain_https_status = models.CharField(max_length=10, verbose_name='子域名HTTPS状态码', null=True, blank=True)
    ip_http_status = models.CharField(max_length=10, verbose_name='IP HTTP状态码', null=True, blank=True)
    ip_https_status = models.CharField(max_length=10, verbose_name='IP HTTPS状态码', null=True, blank=True)

    from_asset = models.CharField(max_length=200, verbose_name='上游资产', null=True, blank=True)

    def __str__(self):
        return f"{self.subdomain} - {self.ip_address} - {self.source}"