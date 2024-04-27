from django.db import models
from common.models import BaseScanJob

class PortScanJob(BaseScanJob):

    @property
    def result_count(self):
        return self.ports.count()  # 返回关联的Subdomain对象的数量

class Port(models.Model):
    scan_job = models.ForeignKey('PortScanJob', on_delete=models.CASCADE, related_name='ports', verbose_name='扫描任务')
    port_number = models.IntegerField(verbose_name='端口号')
    service_name = models.CharField(max_length=100, verbose_name='服务名称', null=True, blank=True)
    protocol = models.CharField(max_length=10, verbose_name='协议', null=True, blank=True)
    ip_address = models.CharField(max_length=15, verbose_name='IP地址')
    state = models.CharField(max_length=20, verbose_name='状态')

    # HTTP相关字段
    http_title = models.CharField(max_length=200, verbose_name='HTTP标题', null=True, blank=True)
    http_code = models.IntegerField(verbose_name='HTTP状态码', null=True, blank=True)

    # HTTPS相关字段
    https_title = models.CharField(max_length=200, verbose_name='HTTPS标题', null=True, blank=True)
    https_code = models.IntegerField(verbose_name='HTTPS状态码', null=True, blank=True)

    @property
    def url(self):
        return f"{self.ip_address}:{self.port_number}"

    def __str__(self):
        return f"{self.ip_address} - {self.port_number}/{self.protocol} - {self.state}"
