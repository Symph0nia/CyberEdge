from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
from .models import Target
import json

@csrf_exempt
@require_http_methods(["GET"])
def get_all_assets_view(request):
    # 获取所有Target实例
    targets = Target.objects.all()
    all_assets = []

    # 遍历每个Target实例，获取相关资产计数
    for target in targets:
        subdomain_count = target.subdomain_count
        port_count = target.port_count
        path_count = target.path_count

        # 构造每个任务的资产信息
        asset_info = {
            'task_id': target.task_id,  # 使用task_id作为唯一标识
            'domain': target.domain,
            'subdomains': subdomain_count,
            'ports': port_count,
            'paths': path_count
        }
        all_assets.append(asset_info)

    # 返回所有任务的资产信息
    return JsonResponse({'assets': all_assets})


@csrf_exempt
@require_http_methods(["POST"])
def create_asset_view(request):
    # 从POST请求的数据中获取'domain'参数
    data = json.loads(request.body.decode('utf-8'))
    domain = data.get('domain')
    if not domain:
        return JsonResponse({'error': 'Missing domain parameter'}, status=400)

    # 检查是否已存在相同域名的Target
    if Target.objects.filter(domain=domain).exists():
        return JsonResponse({'error': 'Domain already exists'}, status=400)

    # 创建并保存新的Target实例
    new_target = Target(domain=domain)
    new_target.save()

    # 构造响应数据
    response_data = {
        'message': 'Asset created successfully'
    }

    return JsonResponse(response_data, status=201)