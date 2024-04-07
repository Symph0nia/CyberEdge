import uuid
from django.db import models

class SubdomainScanJob(models.Model):
    STATUS_CHOICES = [
        ('P', 'Pending'),   # 待处理
        ('R', 'Running'),   # 进行中
        ('C', 'Completed'), # 完成
        ('E', 'Error'),     # 错误
    ]

    task_id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False, verbose_name='任务ID')
    target = models.CharField(max_length=255, verbose_name='扫描目标')
    status = models.CharField(max_length=1, choices=STATUS_CHOICES, default='P', verbose_name='扫描状态')
    start_time = models.DateTimeField(auto_now_add=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误消息')

    def __str__(self):
        return f"{self.target} ({self.get_status_display()})"

class Subdomain(models.Model):
    scan_job = models.ForeignKey(SubdomainScanJob, on_delete=models.CASCADE, related_name='subdomains', verbose_name='子域名扫描任务')
    subdomain = models.CharField(max_length=255, verbose_name='子域名')
    ip_address = models.CharField(max_length=100, verbose_name='IP地址', null=True, blank=True)
    status = models.CharField(max_length=20, verbose_name='状态', null=True, blank=True)
    cname = models.CharField(max_length=255, verbose_name='CNAME', null=True, blank=True)
    port = models.IntegerField(verbose_name='端口', null=True, blank=True)
    title = models.CharField(max_length=255, verbose_name='标题', null=True, blank=True)
    banner = models.CharField(max_length=255, verbose_name='横幅', null=True, blank=True)
    asn = models.CharField(max_length=100, verbose_name='ASN', null=True, blank=True)
    org = models.CharField(max_length=255, verbose_name='组织', null=True, blank=True)
    addr = models.CharField(max_length=255, verbose_name='地址', null=True, blank=True)
    isp = models.CharField(max_length=255, verbose_name='ISP', null=True, blank=True)
    source = models.CharField(max_length=255, verbose_name='来源', null=True, blank=True)
    additional_info = models.TextField(verbose_name='额外信息', null=True, blank=True)  # 用于存储任何额外的扫描信息，现在可以考虑移除或保留用于存储其他信息

    def __str__(self):
        return f"{self.subdomain} - {self.ip_address} - {self.status}"
