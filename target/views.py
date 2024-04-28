from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
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
    # 初始化当前任务的节点
    current_job_node = {
        'name': scan_job.target,  # 任务目标作为节点名称
        'value': scan_job.task_id,  # 任务ID作为节点值
        'children': []  # 初始化子节点列表
    }

    # 获取与当前任务关联的资产列表
    for asset in scan_job.related_assets:
        # 将每个资产作为一个独立的子节点加入
        current_job_node['children'].append({
            'name': asset,  # 资产信息作为节点名称
            'value': scan_job.task_id  # 使用相同的任务ID作为值，因为资产没有独立的task_id
        })

    # 遍历每个子任务，并构建其树状结构
    for child_job in ScanJob.objects.filter(from_job_id=scan_job.task_id):
        current_job_node['children'].append(build_tree_data(child_job))

    # 如果没有子任务和资产，只返回当前任务的基本节点
    return current_job_node if current_job_node['children'] else {
        'name': scan_job.target,
        'value': scan_job.task_id
    }

@csrf_exempt
@require_http_methods(["POST"])
def get_asset_tree_view(request):
    data = json.loads(request.body.decode('utf-8'))
    task_id = data.get('task_id')
    if not task_id:
        return JsonResponse({'error': '域名参数未传递'}, status=400)

    try:
        target = Target.objects.get(task_id=task_id)
    except Target.DoesNotExist:
        return JsonResponse({'error': '域名未设置'}, status=404)

    root_jobs = ScanJob.objects.filter(from_job_id=target.task_id)
    children = [build_tree_data(job) for job in root_jobs]
    tree_data = {
        'name': target.domain,
        'children': children
    }

    return JsonResponse(tree_data)