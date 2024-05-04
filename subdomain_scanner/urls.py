from django.urls import path
from .views import (scan_subdomains_view,
                    subdomain_task_status_view,
                    get_all_tasks_view,
                    delete_task_view,
                    delete_subdomain_view,
                    deduplicate_subdomains_view,
                    delete_http_code_subdomains_view)

urlpatterns = [
    path('scan', scan_subdomains_view, name='scan_subdomain'),
    path('task_status', subdomain_task_status_view, name='subdomain_task_status'),
    path('all_tasks', get_all_tasks_view, name='all_tasks'),
    path('tasks/<uuid:task_id>/delete', delete_task_view, name='delete_task'),
    path('subdomains/<int:id>/delete', delete_subdomain_view, name='delete_subdomain'),
    path('pruning/<uuid:task_id>/duplicate', deduplicate_subdomains_view, name='pruning-duplicate'),
    path('pruning/<uuid:task_id>/status', delete_http_code_subdomains_view, name='pruning-status')
]
