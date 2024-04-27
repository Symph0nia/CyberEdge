import uuid
from django.db import models
from common.models import BaseScanJob
from django.contrib.contenttypes.models import ContentType

class PathScanJob(BaseScanJob):

    from_content_type = models.ForeignKey(
        ContentType,
        on_delete=models.CASCADE,
        null=True,
        blank=True,
        related_name="pathscanjob_from"  # 修改 related_name 使其唯一
    )

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
