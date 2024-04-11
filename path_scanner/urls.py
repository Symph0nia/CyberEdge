from django.urls import path
from .views import scan_paths_view, path_task_status_view, get_all_tasks_view, delete_task_view, delete_path_view

urlpatterns = [
    path('scan', scan_paths_view, name='scan_path'),
    path('task_status', path_task_status_view, name='path_task_status'),
    path('all_tasks', get_all_tasks_view, name='all_tasks'),
    path('tasks/<uuid:task_id>/delete', delete_task_view, name='delete_task'),
    path('paths/<int:id>/delete', delete_path_view, name='delete_path'),
]
