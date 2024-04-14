import uuid
from django.db import models

class PathScanJob(models.Model):
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

    @property
    def result_count(self):
        return self.results.count()

class PathScanResult(models.Model):
    path_scan_job = models.ForeignKey(PathScanJob, on_delete=models.CASCADE, related_name='results', verbose_name='路径扫描任务')
    url = models.URLField(verbose_name='URL')
    content_type = models.CharField(max_length=100, verbose_name='Content-Type', null=True, blank=True)
    status = models.IntegerField(verbose_name='状态码')
    length = models.IntegerField(verbose_name='响应长度')

    def __str__(self):
        return f"{self.url} - Status: {self.status}, Length: {self.length}"
