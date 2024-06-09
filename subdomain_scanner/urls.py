from django.urls import path
from .views import (scan_subdomains_view,
                    subdomain_task_status_view,
                    get_all_tasks_view,
                    delete_task_view,
                    delete_subdomain_view,
                    delete_subdomain_http_ports_view,
                    delete_subdomain_https_ports_view,
                    delete_ip_http_ports_view,
                    delete_ip_https_ports_view
                    )

urlpatterns = [
    path('scan', scan_subdomains_view, name='scan_subdomain'),
    path('task_status', subdomain_task_status_view, name='subdomain_task_status'),
    path('all_tasks', get_all_tasks_view, name='all_tasks'),
    path('tasks/<uuid:task_id>/delete', delete_task_view, name='delete_task'),
    path('subdomains/<int:id>/delete', delete_subdomain_view, name='delete_subdomain'),
    path('pruning/<uuid:task_id>/subdomain_http', delete_subdomain_http_ports_view, name='pruning-subdomain-http-status'),
    path('pruning/<uuid:task_id>/subdomain_https', delete_subdomain_https_ports_view, name='pruning-subdomain-https-status'),
    path('pruning/<uuid:task_id>/ip_http', delete_ip_http_ports_view, name='pruning-ip-http-status'),
    path('pruning/<uuid:task_id>/ip_https', delete_ip_https_ports_view, name='pruning-ip-https-status')
]
