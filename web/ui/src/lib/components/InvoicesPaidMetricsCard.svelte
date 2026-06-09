<script lang="ts">
	import * as Chart from '$lib/components/ui/chart/index.js';

	import { getPaidInvoiceCount, type InvoiceMetric } from '$lib/api/invoices';
	import DashboardCard from './DashboardCard.svelte';
	import Spinner from './Spinner.svelte';
	import { BarChart } from 'layerchart';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';
	import { onDestroy, onMount } from 'svelte';

	let { class: className = '' }: { class?: string } = $props();

	let promise: Promise<InvoiceMetric[]> | null = $state(null);
	let aborter: AbortController | null = null;

	function reload() {
		if (aborter) {
			aborter.abort();
		}

		aborter = new AbortController();

		promise = getPaidInvoiceCount(aborter.signal);
	}

	const chartConfig = {
		paid: { label: 'Created', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;

	onMount(reload);
	onDestroy(() => aborter?.abort());
</script>

<DashboardCard title="Invoices Paid" class={className}>
	<div class="relative h-70 px-6 py-2">
		{#await promise}
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
								bars: {
									stroke: 'transparent'
								},
								xAxis: {
									format: (v: string) =>
										new Date(v + '-01').toLocaleString(undefined, {
											month: 'short'
										})
								},
								highlight: { points: { r: 4 } }
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
				<ErrorMessage variant="warn" text={err.message} retry={reload} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={reload} />
			{/if}
		{/await}
	</div>
</DashboardCard>
