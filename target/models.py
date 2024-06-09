from django.db import models
import uuid

class Target(models.Model):
    TYPE_CHOICES = [
        ('DOMAIN', '域名'),  # 只有一个选项，固定为'DOMAIN'
    ]

    domain = models.CharField(max_length=255, verbose_name='域名')
    task_id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False, verbose_name='任务ID')
    type = models.CharField(max_length=10, choices=TYPE_CHOICES, default='DOMAIN', verbose_name='扫描类型')

    def __str__(self):
        return self.domain

    def count_results_by_type(self, job_type):
        from common.models import ScanJob
        jobs = ScanJob.objects.filter(from_job_id=self.task_id)
        return self._recursive_result_count(jobs, job_type)

    def _recursive_result_count(self, jobs, job_type):
        from common.models import ScanJob
        result_count = 0
        for job in jobs:
            if job.type == job_type:
                result_count += job.result_count  # 累加任务的结果数量
            # 查找此任务下的所有相关任务
            linked_jobs = ScanJob.objects.filter(from_job_id=job.task_id)
            result_count += self._recursive_result_count(linked_jobs, job_type)
        return result_count

    @property
    def subdomain_count(self):
        return self.count_results_by_type('SUBDOMAIN')

    @property
    def port_count(self):
        return self.count_results_by_type('PORT')

    @property
    def path_count(self):
        return self.count_results_by_type('PATH')