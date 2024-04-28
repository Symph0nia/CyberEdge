from django.urls import path
from .views import get_all_assets_view, create_asset_view, get_asset_tree_view

urlpatterns = [
    path('assets', get_all_assets_view, name='assets'),
    path('create', create_asset_view, name='create-target'),
    path('tree', get_asset_tree_view, name='get-tree-data')
]
