<script lang="ts">
	import { SkeletonText, Tag } from 'carbon-components-svelte';
	import type { Trace } from '$lib/types/apiResponse';
	import { FormatTime } from '$lib/types/RFC3339';
	import { StatusMap, PriorityMap } from '$lib/types/enum';

	let {
		t = [],
		loading = false,
		src = 'user'
	}: { t: Trace[]; loading: boolean; src: 'operator' | 'user' | null } = $props();

	function getStatusClass(status?: string) {
		if (status === 'solved' || status === 'escalated') return 'status-green';
		if (status === 'canceled') return 'status-gray';
		if (status === 'fresh') return 'status-red';
		return ''; // 默认为蓝色
	}

	function userOPName(op: string): string {
		if (op === '用户' && src === 'user') {
			return '您';
		}
		return op;
	}
</script>

<div class="timeline-container">
	{#if loading}
		<div class="timeline-skeleton">
			<SkeletonText paragraph lines={3} />
			<br />
			<SkeletonText paragraph lines={3} />
		</div>
	{:else if t.length === 0}
		<div class="empty-state">暂无操作记录</div>
	{:else}
		<ul class="timeline">
			{#each t as trace}
				<li class="timeline-item">
					<!-- 左侧时间点圆圈 -->
					<div class="timeline-marker {getStatusClass(trace.new_status)}"></div>

					<!-- 右侧内容 -->
					<div class="timeline-content">
						<div class="header">
							<span class="operator">{userOPName(trace.op_name)}</span>
							<span class="time">{FormatTime(trace.updated_at)}</span>
						</div>

						<div class="body">
							{#if trace.remark}
								<p class="remark">{trace.remark}</p>
							{/if}

							<!-- 如果有状态/优先级变更，显示标签 -->
							<div class="tags">
								{#if trace.new_status}
									<Tag type="blue" size="sm">{StatusMap[trace.new_status]}</Tag>
								{/if}
								{#if trace.new_priority && src === 'operator'}
									<Tag type="red" size="sm">{PriorityMap[trace.new_priority]}</Tag>
								{/if}
								{#if trace.new_appointment}
									<Tag type="purple" size="sm">预约: {FormatTime(trace.new_appointment)}</Tag>
								{/if}
							</div>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.timeline-container {
		padding: 1rem 0;
		font-family: 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
		width: 100%;
	}

	.empty-state {
		text-align: center;
		color: #8d8d8d;
		padding: 2rem 0;
	}

	.timeline {
		display: flex;
		flex-direction: column;
		list-style: none;
		padding: 0;
		margin: 0;
		position: relative;
		width: 100%;
	}

	/* 垂直连接线 */
	.timeline::before {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		left: 7px; /* 圆点中心对齐 */
		width: 2px;
		background: #e0e0e0;
		z-index: 0;
	}

	.timeline-item {
		display: flex;
		position: relative;
		margin-bottom: 2rem;
		padding-left: 1.5rem; /* 为圆点留出空间 */
		box-sizing: border-box;
		width: 100%;
	}

	.timeline-item:last-child {
		margin-bottom: 0;
	}

	/* 时间轴圆点 */
	.timeline-marker {
		position: absolute;
		left: 0;
		top: 4px;
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: #ffffff;
		border: 4px solid #0f62fe; /* Carbon Blue 60 */
		z-index: 1;
		transition: border-color 0.2s;
	}

	.timeline-marker.status-green {
		border-color: #198038; /* Carbon Green 60 */
	}

	.timeline-marker.status-gray {
		border-color: #8d8d8d; /* Carbon Gray 60 */
	}

	.timeline-marker.status-red {
		border-color: #da1e28; /* Carbon Red 60 */
	}

	.timeline-content {
		position: relative;
		width: 100%;
		min-width: 0;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.25rem;
		width: 100%;
	}

	.operator {
		font-weight: 600;
		font-size: 0.875rem;
		color: #161616;
	}

	.time {
		font-size: 0.75rem;
		color: #6f6f6f;
	}

	.remark {
		font-size: 0.875rem;
		line-height: 1.3;
		color: #393939;
		margin-bottom: 0.5rem;
		white-space: pre-wrap; /* 保留换行符 */
		min-width: 100%;
		word-break: break-word;
		overflow-wrap: break-word;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
</style>
