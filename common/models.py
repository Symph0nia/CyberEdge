from target.models import Target
from django.db import models
import uuid

class ScanJob(models.Model):
    # 定义扫描类型的选择
    TYPE_CHOICES = [
        ('PATH', '路径扫描'),
        ('PORT', '端口扫描'),
        ('SUBDOMAIN', '子域名扫描'),
    ]
    # 定义扫描状态的选择
    STATUS_CHOICES = [
        ('P', 'Pending'),   # 待处理
        ('R', 'Running'),   # 进行中
        ('C', 'Completed'), # 完成
        ('E', 'Error'),     # 错误
    ]

    task_id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False, verbose_name='任务ID')
    target = models.CharField(max_length=255, verbose_name='扫描目标')
    type = models.CharField(max_length=10, choices=TYPE_CHOICES, default='PATH', verbose_name='扫描类型')
    status = models.CharField(max_length=1, choices=STATUS_CHOICES, default='P', verbose_name='扫描状态')
    start_time = models.DateTimeField(auto_now_add=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误消息')
    is_read = models.BooleanField(default=False, verbose_name='是否已读')  # 新增字段，判断是否已读

    from_job_id = models.UUIDField(null=True, blank=True, verbose_name='上游任务ID')

    def __str__(self):
        return f"{self.target} ({self.get_status_display()})"

    @property
    def result_count(self):
        # 根据不同类型返回关联结果的数量
        if self.type == 'PORT':
            return self.ports.count()
        elif self.type == 'PATH':
            return self.paths.count()
        elif self.type == 'SUBDOMAIN':
            return self.subdomains.count()
        return 0

    @property
    def from_job_target(self):
        # 尝试从关联任务或资产中获取目标描述
        if self.from_job_id:
            from_job = ScanJob.objects.filter(task_id=self.from_job_id).first()
            if from_job:
                return f"{from_job.get_type_display()} - {from_job.target}"
            # 模拟一个可能的Target模型的调用，这部分需要根据实际模型来修改
            from_target = Target.objects.filter(task_id=self.from_job_id).first()
            if from_target:
                return f"域名 - {from_target.domain}"
        return None

    @property
    def related_assets(self):
        # 根据任务类型返回相关资产的列表
        assets = {
            'SUBDOMAIN': [{'name': f"子域名:{sub.subdomain}/IP地址:{sub.ip_address}", 'from_asset': sub.from_asset, 'value': sub.id} for sub in self.subdomains.all() if sub.ip_address],
            'PORT': [{'name': f"IP地址:{port.ip_address}/端口:{port.port_number}", 'from_asset': port.from_asset, 'value': port.id} for port in self.ports.all() if port.ip_address],
            'PATH': [{'name': f"路径:{path.path}", 'from_asset': path.from_asset, 'value': path.id} for path in self.paths.all() if path.path],
        }
        return assets.get(self.type, [])
