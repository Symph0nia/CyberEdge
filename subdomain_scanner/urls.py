from django.urls import path
from .views import scan_subdomains_view, subdomain_task_status_view, get_all_tasks_view, delete_task_view, delete_domain_view

urlpatterns = [
    path('scan', scan_subdomains_view, name='scan_subdomain'),
    path('task_status', subdomain_task_status_view, name='subdomain_task_status'),
    path('all_tasks', get_all_tasks_view, name='all_tasks'),
    path('tasks/<uuid:task_id>/delete', delete_task_view, name='delete_task'),
    path('subdomains/<int:id>/delete', delete_domain_view, name='delete_subdomain'),
]
