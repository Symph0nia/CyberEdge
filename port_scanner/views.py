import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .models import Port
from common.models import ScanJob
from .tasks import scan_ports  # 确保正确导入异步任务


@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_ports_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        targets = data.get('target', '')
        ports = data.get('ports', '1-65535')  # 如果未指定，设置默认端口范围
        from_id = data.get('from_id', None)
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 分割逗号分隔的IP地址，并移除空字符串
    targets_list = [ip.strip() for ip in targets.split(',') if ip.strip()]
    # TODO 分割方式和其他的不同，考虑统一
    if not targets_list:
        return JsonResponse({'error': '缺少必要的target参数或格式错误'}, status=400)

    task_ids = []
    # 对每个IP启动一个任务
    for target in targets_list:
        task = scan_ports.delay(target, ports, from_id)
        task_ids.append(task.id)

    # 返回响应
    return JsonResponse({'message': f'共启动{len(task_ids)}个扫描任务', 'task_ids': task_ids})

@csrf_exempt
@require_http_methods(["POST"])
def task_status_view(request):
    try:
        data = json.loads(request.body.decode('utf-8'))  # 解析请求体中的JSON
        task_id = data.get('task_id')
        if not task_id:
            return JsonResponse({'error': '缺少必要的task_id参数'}, status=400)
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    try:
        scan_job = ScanJob.objects.get(task_id=task_id)
    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    # 更新任务为已读状态
    if not scan_job.is_read:
        scan_job.is_read = True
        scan_job.save()

    response_data = {
        'task_id': task_id,
        'task_status': scan_job.get_status_display(),
    }

    if scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'ports': [
                {
                    'id': port.id,
                    'url': port.url,
                    'ip_address': port.ip_address,
                    'port_number': port.port_number,
                    'service_name': port.service_name,
                    'protocol': port.protocol,
                    'state': port.state,
                    'http_code': port.http_code,
                    'http_title': port.http_title,
                    'https_code': port.https_code,
                    'https_title': port.https_title,
                    'from_asset': port.from_asset,
                } for port in scan_job.ports.all()  # 假设ports是与ScanJob相关联的模型实例
            ],
            'error_message': scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])  # 修改为接受GET请求
def get_all_tasks_view(request):
    # 获取所有ScanJob实例的概要信息
    tasks = ScanJob.objects.filter(type='PORT')
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
def delete_port_view(request, id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = Port.objects.get(id=id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '端口删除成功'}, status=200)
    except Port.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '端口ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除端口时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_http_ports_view(request, task_id):
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
        specific_http_ports = Port.objects.filter(scan_job_id=task_id, http_code=status_code)

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
def delete_https_ports_view(request, task_id):
    try:
        # 从请求中获取 http_code 参数
        status_code = request.GET.get('status_code')
        if not status_code:
            return JsonResponse({'error': '缺少必要的 status_code 参数'}, status=400)

        # 将 http_code 转换为整数
        try:
            http_code = int(status_code)
        except ValueError:
            return JsonResponse({'error': 'status_code 参数必须是整数'}, status=400)

        # 获取指定ScanJob的所有端口记录，其HTTP状态码等于指定的http_code
        specific_http_ports = Port.objects.filter(scan_job_id=task_id, https_code=status_code)

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