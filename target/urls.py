from django.urls import path
from .views import get_all_assets_view, create_asset_view

urlpatterns = [
    path('assets', get_all_assets_view, name='assets'),
    path('create', create_asset_view, name='create-target'),
]
