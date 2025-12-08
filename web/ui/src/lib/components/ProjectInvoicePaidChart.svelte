<script lang="ts">
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { curveLinear } from 'd3-shape';
	import { scalePoint } from 'd3-scale';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getPaidInvoiceCountByProject, type InvoiceMetric } from '$lib/api/invoices';
	import { onMount } from 'svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import Spinner from './Spinner.svelte';
	import { LineChart } from 'layerchart';

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

<div
	class="grid h-full w-full grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
>
	<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Invoices Paid</p>
	</div>
	<div class="h-70 relative px-6 py-2">
		{#await metricsPromise}
			<Spinner />
		{:then res}
			{#if res}
				<div class="absolute inset-x-6 inset-y-2">
					<Chart.Container config={chartConfig} class="h-full w-full">
						<LineChart
							data={res}
							x="key"
							xScale={scalePoint()}
							yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
							axis="x"
							series={[
								{
									key: 'value',
									label: chartConfig.paid.label,
									color: chartConfig.paid.color
								}
							]}
							props={{
								spline: { curve: curveLinear, motion: 'none', strokeWidth: 2 },
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
						</LineChart>
					</Chart.Container>
				</div>
			{:else}
				<ErrorMessage variant="warn" text="Metrics Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadCount} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Metrics"
					retry={loadCount}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadCount} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadCount} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadCount} />
			{/if}
		{/await}
	</div>
</div>
