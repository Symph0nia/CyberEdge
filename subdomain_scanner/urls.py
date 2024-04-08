from django.urls import path
from .views import scan_subdomains_view, subdomain_task_status_view

urlpatterns = [
    path('scan', scan_subdomains_view, name='scan_subdomain'),
    path('task_status', subdomain_task_status_view, name='subdomain_task_status'),
]
