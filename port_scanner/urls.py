from django.urls import path
from .views import (scan_ports_view,
                    task_status_view,
                    get_all_tasks_view,
                    delete_task_view,
                    delete_port_view,
                    delete_http_ports_view,
                    delete_https_ports_view,
                    )

urlpatterns = [
    path('scan', scan_ports_view, name='scan_ports'),
    path('task_status', task_status_view, name='task_status'),
    path('all_tasks', get_all_tasks_view, name='all_tasks'),
    path('tasks/<uuid:task_id>/delete', delete_task_view, name='delete_task'),
    path('ports/<int:id>/delete', delete_port_view, name='delete_port'),
    path('pruning/<uuid:task_id>/http', delete_http_ports_view, name='pruning-http'),
    path('pruning/<uuid:task_id>/https', delete_https_ports_view, name='pruning-https')
]
