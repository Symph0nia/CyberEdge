from django.db import models
from common.models import ScanJob
from django.contrib.contenttypes.models import ContentType

class Path(models.Model):
    scan_job = models.ForeignKey(ScanJob, on_delete=models.CASCADE, related_name='paths', verbose_name='路径扫描任务')
    url = models.URLField(verbose_name='URL')
    path = models.URLField(verbose_name='路径')
    content_type = models.CharField(max_length=100, verbose_name='Content-Type', null=True, blank=True)
    status = models.IntegerField(verbose_name='状态码')
    length = models.IntegerField(verbose_name='响应长度')

    from_asset = models.CharField(max_length=200, verbose_name='上游资产', null=True, blank=True)

    def __str__(self):
        return f"{self.url} - Status: {self.status}, Length: {self.length}"