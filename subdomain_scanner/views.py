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
            'subdomains': list(subdomain_scan_job.subdomains.values(
                'id',
                'subdomain',
                'ip_address',
                'source',  # 新增源字段
                'subdomain_http_status',  # 新增子域名HTTP状态码字段
                'subdomain_https_status',  # 新增子域名HTTPS状态码字段
                'ip_http_status',  # 新增IP HTTP状态码字段
                'ip_https_status',  # 新增IP HTTPS状态码字段
                'from_asset',  # 保留上游资产字段
            )),
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
@require_http_methods(["DELETE"])
def delete_subdomain_http_ports_view(request, task_id):
    try:
        # 从请求中获取 http_code 参数
        status_code = request.GET.get('status_code')
        if not status_code:
            return JsonResponse({'error': '缺少必要的 status_code 参数'}, status=400)

        # 将status_code转换为整数
        try:
            status_code = int(status_code)
        except ValueError:
            return JsonResponse({'error': 'status_code参数必须是整数'}, status=400)

        # 获取指定ScanJob的所有端口记录，其HTTP状态码等于指定的http_code
        specific_http_ports = Subdomain.objects.filter(scan_job_id=task_id, subdomain_http_status=status_code)

        # 记录将要删除的记录数量
        count_to_delete = specific_http_ports.count()

        # 删除这些记录
        specific_http_ports.delete()

        return JsonResponse({
            'message': f'成功删除{count_to_delete}个HTTP状态码为{status_code}的端口。',
            'deleted': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法执行删除'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'删除操作时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_subdomain_https_ports_view(request, task_id):
    try:
        # 从请求中获取 http_code 参数
        status_code = request.GET.get('status_code')
        if not status_code:
            return JsonResponse({'error': '缺少必要的 status_code 参数'}, status=400)

        # 将status_code转换为整数
        try:
            status_code = int(status_code)
        except ValueError:
            return JsonResponse({'error': 'status_code参数必须是整数'}, status=400)

        # 获取指定ScanJob的所有端口记录，其HTTP状态码等于指定的http_code
        specific_https_ports = Subdomain.objects.filter(scan_job_id=task_id, subdomain_https_status=status_code)

        # 记录将要删除的记录数量
        count_to_delete = specific_https_ports.count()

        # 删除这些记录
        specific_https_ports.delete()

        return JsonResponse({
            'message': f'成功删除{count_to_delete}个HTTP状态码为{status_code}的端口。',
            'deleted': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法执行删除'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'删除操作时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_ip_http_ports_view(request, task_id):
    try:
        # 从请求中获取 http_code 参数
        status_code = request.GET.get('status_code')
        if not status_code:
            return JsonResponse({'error': '缺少必要的 status_code 参数'}, status=400)

        # 将status_code转换为整数
        try:
            status_code = int(status_code)
        except ValueError:
            return JsonResponse({'error': 'status_code参数必须是整数'}, status=400)

        # 获取指定ScanJob的所有端口记录，其HTTP状态码等于指定的http_code
        specific_http_ports = Subdomain.objects.filter(scan_job_id=task_id, ip_http_status=status_code)

        # 记录将要删除的记录数量
        count_to_delete = specific_http_ports.count()

        # 删除这些记录
        specific_http_ports.delete()

        return JsonResponse({
            'message': f'成功删除{count_to_delete}个HTTP状态码为{status_code}的端口。',
            'deleted': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法执行删除'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'删除操作时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_ip_https_ports_view(request, task_id):
    try:
        # 从请求中获取 http_code 参数
        status_code = request.GET.get('status_code')
        if not status_code:
            return JsonResponse({'error': '缺少必要的 status_code 参数'}, status=400)

        # 将status_code转换为整数
        try:
            status_code = int(status_code)
        except ValueError:
            return JsonResponse({'error': 'status_code参数必须是整数'}, status=400)

        # 获取指定ScanJob的所有端口记录，其HTTP状态码等于指定的http_code
        specific_https_ports = Subdomain.objects.filter(scan_job_id=task_id, ip_https_status=status_code)

        # 记录将要删除的记录数量
        count_to_delete = specific_https_ports.count()

        # 删除这些记录
        specific_https_ports.delete()

        return JsonResponse({
            'message': f'成功删除{count_to_delete}个HTTP状态码为{status_code}的端口。',
            'deleted': True
        }, status=200)

    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '指定的ScanJob不存在，无法执行删除'}, status=404)
    except Exception as e:
        return JsonResponse({'error': f'删除操作时发生错误: {str(e)}'}, status=500)