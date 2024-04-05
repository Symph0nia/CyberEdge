from django.db import models
import uuid

class ScanJob(models.Model):
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

class Port(models.Model):
    scan_job = models.ForeignKey(ScanJob, on_delete=models.CASCADE, related_name='ports', verbose_name='扫描任务')
    port_number = models.IntegerField(verbose_name='端口号')
    service_name = models.CharField(max_length=100, verbose_name='服务名称', null=True, blank=True)
    protocol = models.CharField(max_length=10, verbose_name='协议', null=True, blank=True)
    state = models.CharField(max_length=20, verbose_name='状态')

    def __str__(self):
        return f"{self.port_number}/{self.protocol} - {self.state}"
