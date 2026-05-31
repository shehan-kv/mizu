<script lang="ts">
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { BarChart } from 'layerchart';
	import Spinner from './Spinner.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { onMount } from 'svelte';
	import { getTaskCompleteCountByProject, type TaskMetric } from '$lib/api/task';
	import { ApiError } from '$lib/api/client';

	interface Props {
		projectId: string;
	}
	let { projectId }: Props = $props();

	let metricsPromise: Promise<TaskMetric[]> | null = $state(null);
	let abort: AbortController | null = null;
	function loadMetrics() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		metricsPromise = getTaskCompleteCountByProject(projectId, abort.signal);
	}

	export function refresh() {
		loadMetrics();
	}

	onMount(() => {
		loadMetrics();
	});

	const chartConfig = {
		completed: { label: 'Completed', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden rounded border">
	<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Tasks Completed</p>
	</div>
	<div class="px-6 py-2">
		{#await metricsPromise}
			<Spinner />
		{:then res}
			{#if res && res.length > 0}
				<Chart.Container config={chartConfig} class="h-70 w-full">
					<BarChart
						data={res}
						x="key"
						axis="x"
						yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
						seriesLayout="group"
						series={[
							{
								key: 'value',
								label: chartConfig.completed.label,
								color: chartConfig.completed.color
							}
						]}
						props={{
							xAxis: {
								format: (v: string) =>
									new Date(v).toLocaleString(undefined, {
										month: 'short',
										day: '2-digit'
									})
							},
							bars: {
								stroke: 'transparent'
							}
						}}
					>
						{#snippet tooltip()}
							<Chart.Tooltip />
						{/snippet}
					</BarChart>
				</Chart.Container>
			{:else}
				<ErrorMessage variant="warn" text="Metrics Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadMetrics} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadMetrics} />
			{/if}
		{/await}
	</div>
</div>
