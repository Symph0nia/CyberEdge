import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
from django.db.models import Min

from .models import Subdomain
from common.models import ScanJob
from .tasks import scan_subdomains  # 确保从你的Celery任务模块导入scan_subdomains函数


@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])
def scan_subdomains_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        targets = data.get('targets', [])  # 从请求中获取目标域名列表
        from_id = data.get('from_id', None)
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 确保targets是一个列表
    if not isinstance(targets, list) or not all(isinstance(target, str) for target in targets):
        return JsonResponse({'error': 'targets参数必须是字符串的数组'}, status=400)

    if not targets:
        return JsonResponse({'error': '缺少必要的targets参数'}, status=400)

    task_ids = []
    # 对每个目标域名启动一个子域名扫描任务
    for target in targets:
        # 假设scan_subdomains.delay是一个异步任务启动函数
        task = scan_subdomains.delay(target, from_id)
        task_ids.append(task.id)

    # 返回响应
    return JsonResponse({'message': f'共启动{len(task_ids)}个子域名扫描任务', 'task_ids': task_ids})

@csrf_exempt
@require_http_methods(["POST"])
def subdomain_task_status_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        task_id = data.get('task_id')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not task_id:
        return JsonResponse({'error': '缺少必要的task_id参数'}, status=400)

    try:
        subdomain_scan_job = ScanJob.objects.get(task_id=task_id)
    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    if not subdomain_scan_job.is_read:
        subdomain_scan_job.is_read = True
        subdomain_scan_job.save()

    # 构造响应数据
    response_data = {
        'task_id': task_id,
        'task_status': subdomain_scan_job.get_status_display(),
    }

    if subdomain_scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'subdomains': list(subdomain_scan_job.subdomains.values('id', 'subdomain', 'ip_address', 'status', 'cname', 'port', 'title',
                    'banner', 'addr', 'from_asset')),
            'error_message': subdomain_scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])  # 修改为接受GET请求
def get_all_tasks_view(request):
    # 获取所有ScanJob实例的概要信息
    tasks = ScanJob.objects.filter(type='SUBDOMAIN')
    tasks_list = []
    for task in tasks:
        tasks_list.append({
            'task_id': task.task_id,
            'target': task.target,
            'status': task.status,
            'result_count': task.result_count,
            'start_time': task.start_time.strftime('%Y年%m月%d日 %H:%M:%S') if task.start_time else None,
            'end_time': task.end_time.strftime('%Y年%m月%d日 %H:%M:%S') if task.end_time else None,
            'from': task.from_job_target,
            'is_read': task.is_read,
        })

    # 返回响应
    return JsonResponse({'tasks': tasks_list}, safe=False)  # safe=False允许非字典对象被序列化为JSON

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_task_view(request, task_id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = ScanJob.objects.get(task_id=task_id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '任务删除成功'}, status=200)
    except ScanJob.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '任务ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除任务时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_subdomain_view(request, id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = Subdomain.objects.get(id=id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '子域名删除成功'}, status=200)
    except Subdomain.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '子域名ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除子域名时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["POST"])
def deduplicate_subdomains_view(request, task_id):
    try:
        subdomains = Subdomain.objects.filter(scan_job_id=task_id)
        # 对于每个subdomain，找到最小的id值（即最早的记录）
        min_ids = subdomains.values('subdomain').annotate(min_id=Min('id'))

        # 构建一个包含所有最小id的列表，这些是将要保留的记录
        min_id_list = [item['min_id'] for item in min_ids]

        # 删除那些id不在min_id_list中的所有记录
        deleted_count = subdomains.exclude(id__in=min_id_list).delete()[0]  # delete返回一个元组，第一个元素是删除的计数

        return JsonResponse({
            'message': f'成功删除{deleted_count}个重复的子域名。',
            'deduplicated': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法进行去重'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'去重操作时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_http_code_subdomains_view(request, task_id):
    try:
        # 获取指定ScanJob的所有子域名，其状态非'200'
        non_200_subdomains = Subdomain.objects.filter(scan_job_id=task_id).exclude(status='200')

        # 记录将要删除的记录数量
        count_to_delete = non_200_subdomains.count()

        # 删除这些记录
        non_200_subdomains.delete()

        return JsonResponse({
            'message': f'成功删除{count_to_delete}个状态非200的子域名。',
            'deleted': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法执行删除'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'删除操作时发生错误: {str(e)}'}, status=500)
