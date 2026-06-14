<script lang="ts">
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { getPaidInvoiceCountByProject, type InvoiceMetric } from '$lib/api/invoices';
	import { onMount } from 'svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import Spinner from './Spinner.svelte';
	import { BarChart } from 'layerchart';
	import { ApiError } from '$lib/api/client';

	let { projectId } = $props();

	let metricsPromise: Promise<InvoiceMetric[]> | null = $state(null);
	let abort: AbortController | null = null;
	function loadCount() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		metricsPromise = getPaidInvoiceCountByProject(projectId, abort.signal);
	}

	onMount(() => {
		loadCount();
	});

	const chartConfig = {
		paid: { label: 'Paid', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden">
	<div class="border-b px-6 py-2">
		<p class="text-sm">Invoices Paid</p>
	</div>
	<div class="relative h-70 px-6 py-2">
		{#await metricsPromise}
			<Spinner />
		{:then res}
			{#if res}
				<div class="absolute inset-x-6 inset-y-2">
					<Chart.Container config={chartConfig} class="h-full w-full">
						<BarChart
							data={res}
							x="key"
							axis="x"
							yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
							seriesLayout="group"
							series={[
								{
									key: 'value',
									label: chartConfig.paid.label,
									color: chartConfig.paid.color
								}
							]}
							props={{
								xAxis: {
									format: (v: string) =>
										new Date(v).toLocaleString(undefined, {
											month: 'short'
										})
								},
								bars: {
									stroke: 'transparent'
								}
							}}
						>
							{#snippet tooltip()}
								<Chart.Tooltip hideLabel />
							{/snippet}
						</BarChart>
					</Chart.Container>
				</div>
			{:else}
				<ErrorMessage variant="warn" text="Metrics Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadCount} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadCount} />
			{/if}
		{/await}
	</div>
</div>
