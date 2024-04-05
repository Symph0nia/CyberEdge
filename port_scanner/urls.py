from django.urls import path
from .views import scan_ports_view, task_status_view

urlpatterns = [
    path('scan/', scan_ports_view, name='scan_ports'),
    path('task_status/', task_status_view, name='task_status'),
]
