from django.db import models
from django.contrib.contenttypes.fields import GenericForeignKey
from django.contrib.contenttypes.models import ContentType

class Target(models.Model):
    domain = models.CharField(max_length=255, verbose_name='域名')

    from_content_type = models.ForeignKey(ContentType, on_delete=models.CASCADE, null=True, blank=True,
                                          related_name="base_scanjob_from")
    from_object_id = models.UUIDField(null=True, blank=True)
    from_job = GenericForeignKey('from_content_type', 'from_object_id')

    def __str__(self):
        return self.domain

    @property
    def subdomain_count(self):
        # 假设存在 Subdomain 模型，并有外键指向 Target
        return self.subdomain_set.count()

    @property
    def ip_count(self):
        # 假设存在 IpAddress 模型，并有外键指向 Target
        return self.ipaddress_set.count()

    @property
    def port_count(self):
        # 假设存在 Port 模型，并有外键指向 Target
        return self.port_set.count()

    @property
    def path_count(self):
        # 假设存在 Path 模型，并有外键指向 Target
        return self.path_set.count()