from django.urls import path
from .views import full_scan_view

urlpatterns = [
    path('scan', full_scan_view, name='scan_full')
]
