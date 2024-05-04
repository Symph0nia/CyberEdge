from django.db import models
from common.models import ScanJob
from django.contrib.contenttypes.models import ContentType

class Subdomain(models.Model):
    scan_job = models.ForeignKey(ScanJob, on_delete=models.CASCADE, related_name='subdomains', verbose_name='子域名扫描任务')
    subdomain = models.CharField(max_length=255, verbose_name='子域名')
    domain = models.CharField(max_length=255, verbose_name='域名', blank=True, editable=False)
    ip_address = models.CharField(max_length=100, verbose_name='IP地址', null=True, blank=True)
    status = models.CharField(max_length=20, verbose_name='状态', null=True, blank=True)
    cname = models.CharField(max_length=255, verbose_name='CNAME', null=True, blank=True)
    port = models.IntegerField(verbose_name='端口', null=True, blank=True)
    title = models.CharField(max_length=255, verbose_name='标题', null=True, blank=True)
    banner = models.CharField(max_length=255, verbose_name='横幅', null=True, blank=True)
    addr = models.CharField(max_length=255, verbose_name='地址', null=True, blank=True)

    from_asset = models.CharField(max_length=200, verbose_name='上游资产', null=True, blank=True)

    def __str__(self):
        return f"{self.subdomain} - {self.ip_address} - {self.status}"