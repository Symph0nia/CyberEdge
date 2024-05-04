from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
from django.db.models import Q
from .models import Target
from common.models import ScanJob
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
        return JsonResponse({'error': '域名参数未传递'}, status=400)

    # 检查是否已存在相同域名的Target
    if Target.objects.filter(domain=domain).exists():
        return JsonResponse({'error': '域名不存在'}, status=400)

    # 创建并保存新的Target实例
    new_target = Target(domain=domain)
    new_target.save()

    # 构造响应数据
    response_data = {
        'message': 'Asset created successfully'
    }

    return JsonResponse(response_data, status=201)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_target_view(request, task_id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = Target.objects.get(task_id=task_id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '目标删除成功'}, status=200)
    except Target.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '目标ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除目标时发生错误: {str(e)}'}, status=500)


def build_tree_data(scan_job):
    nodes = {}
    root_node = {
        'name': scan_job.target,
        'value': scan_job.task_id.hex,
        'children': []
    }
    nodes[scan_job.task_id.hex] = root_node

    # 处理资产并建立父子关系
    for asset_type in ['subdomains', 'ports', 'paths']:
        for asset in getattr(scan_job, asset_type).all():
            asset_node = {
                'name': asset.__str__(),
                'value': asset.id,
                'children': []
            }
            nodes[asset.id] = asset_node
            # 使用 from_asset 建立父子关系
            parent_value = asset.from_asset
            if parent_value and parent_value in nodes:
                nodes[parent_value]['children'].append(asset_node)
            elif parent_value:
                # 如果父资产在 nodes 中不存在但存在 from_asset，创建一个新的根节点
                parent_node = {
                    'name': parent_value,
                    'value': parent_value,  # 这里我们假设 parent_value 唯一
                    'children': [asset_node]
                }
                nodes[parent_value] = parent_node
                root_node['children'].append(parent_node)
            else:
                # 附加到根节点
                root_node['children'].append(asset_node)

    # 递归处理子任务
    for child_job in ScanJob.objects.filter(from_job_id=scan_job.task_id):
        child_node = build_tree_data(child_job)
        # 仅在子节点还未被添加时，才将其添加到树中
        if child_node['value'] not in nodes:
            nodes[child_node['value']] = child_node
            root_node['children'].append(child_node)

    return root_node


@csrf_exempt
@require_http_methods(["POST"])
def get_asset_tree_view(request):
    data = json.loads(request.body.decode('utf-8'))
    task_id = data.get('task_id')
    if not task_id:
        return JsonResponse({'error': 'Task ID not provided'}, status=400)

    try:
        target = Target.objects.get(task_id=task_id)
    except Target.DoesNotExist:
        return JsonResponse({'error': 'Target not found'}, status=404)

    # 使用 target.domain 从 ScanJob 的相关资产模型中查询数据
    root_jobs = ScanJob.objects.filter(
        Q(subdomains__from_asset=target.domain) |
        Q(ports__from_asset=target.domain) |
        Q(paths__from_asset=target.domain)
    ).distinct()
    children = [build_tree_data(job) for job in root_jobs]
    tree_data = {
        'name': target.domain,
        'children': children
    }

    return JsonResponse(tree_data)