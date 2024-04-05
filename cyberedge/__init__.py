from __future__ import absolute_import, unicode_literals

# 这将确保应用在 Django 启动时就被加载
from .celery import app as celery_app

__all__ = ('celery_app',)
