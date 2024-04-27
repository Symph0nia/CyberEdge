from django.db import models
import uuid

class BaseScanJob(models.Model):
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
    from_job = models.UUIDField(null=True, blank=True, verbose_name='上游任务ID')
    to_job = models.UUIDField(null=True, blank=True, verbose_name='下游任务ID')

    class Meta:
        abstract = True  # 设定这个类为抽象类，不会创建数据库表

    def __str__(self):
        return f"{self.target} ({self.get_status_display()})"