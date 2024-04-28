from django.contrib.contenttypes.fields import GenericForeignKey
from django.contrib.contenttypes.models import ContentType
from django.db import models
import uuid

class ScanJob(models.Model):
    TYPE_CHOICES = [
        ('PATH', '路径扫描'),
        ('PORT', '端口扫描'),
        ('SUBDOMAIN', '子域名扫描'),
    ]

    task_id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False, verbose_name='任务ID')
    target = models.CharField(max_length=255, verbose_name='扫描目标')
    type = models.CharField(max_length=10, choices=TYPE_CHOICES, default='PATH', verbose_name='扫描类型')
    status = models.CharField(max_length=1, choices=[
        ('P', 'Pending'),   # 待处理
        ('R', 'Running'),   # 进行中
        ('C', 'Completed'), # 完成
        ('E', 'Error'),     # 错误
    ], default='P', verbose_name='扫描状态')
    start_time = models.DateTimeField(auto_now_add=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误消息')

    from_job_id = models.UUIDField(null=True, blank=True, verbose_name='上游任务ID')

    def __str__(self):
        return f"{self.target} ({self.get_status_display()})"

    @property
    def result_count(self):
        if self.type == 'PORT':
            return self.ports.count()  # 假设ports是关联的端口扫描结果
        elif self.type == 'PATH':
            return self.paths.count()  # 假设results是关联的路径扫描结果
        elif self.type == 'SUBDOMAIN':
            return self.subdomains.count()  # 假设results是关联的路径扫描结果
        else:
            return 0  # 适当地处理其他类型或当没有结果时

    @property
    def from_job_target(self):
        if self.from_job_id:
            try:
                from_job = ScanJob.objects.get(task_id=self.from_job_id)
                return from_job.target
            except ScanJob.DoesNotExist:
                return None
        else:
            return None