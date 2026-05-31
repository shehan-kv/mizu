<script lang="ts">
	import {
		getProjectCreatedMetrics,
		getProjectStats,
		type ProjectMetric,
		type ProjectStat
	} from '$lib/api/projects';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import * as Table from '$lib/components/ui/table';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { BarChart } from 'layerchart';
	import { onMount } from 'svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { formatDate } from '$lib/utils/formatDate';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { getContractOverviews, type Contract } from '$lib/api/contracts';
	import {
		getInvoices,
		getInvoiceSummmary,
		getPaidInvoiceCount,
		type InvoiceMetric,
		type InvoiceOverview,
		type InvoicesSummary,
		type InvoicesSummaryMetric
	} from '$lib/api/invoices';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDecimalSuffix } from '$lib/utils/formatDecimalSuffix';
	import type { PaginatedResponse } from '$lib/api/page';
	import { resolve } from '$app/paths';
	import { ApiError } from '$lib/api/client';
	import DashboardCard from '$lib/components/DashboardCard.svelte';

	let projectsPromise: Promise<PaginatedResponse<ProjectStat>> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProjects() {
		if (projectAbort) {
			projectAbort.abort();
		}
		projectAbort = new AbortController();

		projectsPromise = getProjectStats({ page: 1, limit: 20 }, projectAbort.signal);
	}

	let projectMetricsPromise: Promise<ProjectMetric[]> | null = $state(null);
	let projectMetricsAbort: AbortController | null = null;
	function loadProjectCountMetrics() {
		if (projectMetricsAbort) {
			projectMetricsAbort.abort();
		}
		projectMetricsAbort = new AbortController();

		projectMetricsPromise = getProjectCreatedMetrics(projectMetricsAbort.signal);
	}

	let contractsPromise: Promise<PaginatedResponse<Contract>> | null = $state(null);
	let contractsAbort: AbortController | null = null;
	function loadContracts() {
		if (contractsAbort) {
			contractsAbort.abort();
		}

		contractsAbort = new AbortController();

		contractsPromise = getContractOverviews({ page: 1, limit: 20 }, contractsAbort.signal);
	}

	let invoicesPromise: Promise<PaginatedResponse<InvoiceOverview>> | null = $state(null);
	let invoicesAbort: AbortController | null = null;
	function loadInvoices() {
		if (invoicesAbort) {
			invoicesAbort.abort();
		}

		invoicesAbort = new AbortController();

		invoicesPromise = getInvoices({ page: 1, limit: 20 }, invoicesAbort.signal);
	}

	let invoiceMetricsPromise: Promise<InvoiceMetric[]> | null = $state(null);
	let invoiceMetricsAbort: AbortController | null = null;
	function loadInvoiceMetrics() {
		if (invoiceMetricsAbort) {
			invoiceMetricsAbort.abort();
		}

		invoiceMetricsAbort = new AbortController();

		invoiceMetricsPromise = getPaidInvoiceCount(invoiceMetricsAbort.signal);
	}

	let invoiceOverviewPromise: Promise<InvoicesSummary> | null = $state(null);
	let invoiceOverviewAbort: AbortController | null = null;
	function loadInvoiceOverview() {
		if (invoiceOverviewAbort) {
			invoiceOverviewAbort.abort();
		}

		invoiceOverviewAbort = new AbortController();

		invoiceOverviewPromise = getInvoiceSummmary(invoiceOverviewAbort.signal);
	}

	const invChartConfig = {
		paid: { label: 'Created', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;

	onMount(() => {
		loadProjectCountMetrics();
		loadProjects();
		loadContracts();
		loadInvoices();
		loadInvoiceMetrics();
		loadInvoiceOverview();
	});
</script>

<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<DashboardCard title="Projects Created" class="col-span-2 min-h-80">
		<div class="relative px-6 py-2">
			{#await projectMetricsPromise}
				<Spinner />
			{:then res}
				{#if res}
					<div class="absolute inset-x-6 inset-y-2">
						<Chart.Container config={invChartConfig} class="h-full w-full">
							<BarChart
								data={res}
								x="key"
								axis="x"
								yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
								seriesLayout="group"
								series={[
									{
										key: 'value',
										label: invChartConfig.paid.label,
										color: invChartConfig.paid.color
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
					<ErrorMessage variant="warn" text={err.message} retry={loadProjectCountMetrics} />
				{:else}
					<ErrorMessage variant="warn" text="An Error Occurred" retry={loadProjectCountMetrics} />
				{/if}
			{/await}
		</div>
	</DashboardCard>

	<DashboardCard title="Recent Projects" class="col-span-2 min-h-80">
		<div class="overflow-scroll px-6 py-2">
			{#await projectsPromise}
				<Spinner />
			{:then res}
				{#if res && res.items.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.items as project (project.id)}
								<Table.Row
									class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
							dark:text-neutral-400 dark:hover:text-neutral-50"
								>
									<Table.Cell class="pl-0">{project.name}</Table.Cell>
									<Table.Cell class="flex items-center gap-1.5">
										{#if project.status == 'started'}
											<span class="relative flex size-2">
												<span
													class="absolute inline-flex h-full w-full animate-ping rounded-full
									bg-green-500 opacity-75 dark:bg-green-600"
												>
												</span>
												<span
													class="relative inline-flex size-2 rounded-full bg-green-500 dark:bg-green-600"
												></span>
											</span>
										{/if}
										{toTitleCase(project.status)}
									</Table.Cell>
									<Table.Cell class="pr-0" align="right">
										<a href={resolve(`/projects/${project.id}`)} title="View">
											<ArrowRight size={18} />
										</a>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{:else}
					<ErrorMessage variant="warn" text="Projects Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof ApiError}
					<ErrorMessage variant="warn" text={err.message} retry={loadProjects} />
				{:else}
					<ErrorMessage variant="warn" text="An Error Occurred" retry={loadProjects} />
				{/if}
			{/await}
		</div>
	</DashboardCard>

	<DashboardCard title="Recent Contracts" class="col-span-4 max-h-100 min-h-80">
		<div class="overflow-scroll px-6 py-2">
			{#await contractsPromise}
				<Spinner />
			{:then res}
				{#if res && res.items.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.items as contract (contract.id)}
								<Table.Row
									class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
								>
									<Table.Cell class="pl-0">
										{contract.name}
									</Table.Cell>
									<Table.Cell class="flex items-center gap-1">
										{toTitleCase(contract.status)}
										{#if contract.status == 'signed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</Table.Cell>
									<Table.Cell>
										Created On {formatDate(contract.createdAt)}
									</Table.Cell>
									<Table.Cell class="pr-0" align="right">
										<a href={resolve(`/contracts/${contract.id}`)} title="View">
											<ArrowRight size={18} />
										</a>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{:else}
					<ErrorMessage variant="warn" text="Contracts Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof ApiError}
					<ErrorMessage variant="warn" text={err.message} retry={loadContracts} />
				{:else}
					<ErrorMessage variant="warn" text="An Error Occurred" retry={loadContracts} />
				{/if}
			{/await}
		</div>
	</DashboardCard>

	<div class="col-span-4 grid grid-cols-5 gap-2">
		<DashboardCard title="Recent Invoices" class="col-span-2 max-h-80 min-h-50">
			<div class="overflow-scroll px-6 py-2">
				{#await invoicesPromise}
					<Spinner />
				{:then res}
					{#if res && res.items.length > 0}
						<Table.Root>
							<Table.Body>
								{#each res.items as invoice (invoice.id)}
									<Table.Row
										class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
									>
										<Table.Cell class="pl-0">
											{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id}
										</Table.Cell>
										<Table.Cell align="right">
											{currencyFormatter(invoice.currencyCode, invoice.subTotal)} Total
										</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(invoice.status)}
											{#if invoice.status == 'paid' || invoice.status == 'accepted'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell class="pr-0" align="right">
											<a href={resolve(`/invoices-and-quotes/${invoice.id}`)} title="View">
												<ArrowRight size={18} />
											</a>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{:else}
						<ErrorMessage variant="warn" text="Invoices/Quotes Not Found" />
					{/if}
				{:catch err}
					{#if err instanceof ApiError}
						<ErrorMessage variant="warn" text={err.message} retry={loadInvoices} />
					{:else}
						<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoices} />
					{/if}
				{/await}
			</div>
		</DashboardCard>

		<DashboardCard title="Invoices Overview" class="max-h-80 min-h-50">
			<div class="space-y-4 overflow-scroll px-6 py-2">
				{#snippet overview(label: string, items: InvoicesSummaryMetric[])}
					<div>
						<p class="border-b py-2 text-sm text-neutral-600 dark:text-neutral-400">
							{label}
						</p>
						{#if items.length > 0}
							{#each items as item (item)}
								<p
									class="py-2 text-right text-3xl"
									title={currencyFormatter(item.currencyCode, item.amount)}
								>
									{formatDecimalSuffix(item.amount)}
									<span class="text-sm">
										{item.currencyCode.toUpperCase()}

										({item.count})
									</span>
								</p>
							{/each}
						{:else}
							<p class="py-2 text-right text-2xl">N/A</p>
						{/if}
					</div>
				{/snippet}

				{#await invoiceOverviewPromise}
					<Spinner />
				{:then res}
					{#if res}
						{@render overview('Invoices Paid', res.paid)}
						{@render overview('Invoices Pending', res.pending)}
						{@render overview('Invoices Accepted', res.accepted)}
						{@render overview('Invoices Rejected', res.rejected)}
						{@render overview('Invoices Cancelled', res.cancelled)}
						{@render overview('Quotes Pending', res.quotesPending)}
						{@render overview('Quotes Rejected', res.quotesRejected)}
					{:else}
						<ErrorMessage variant="warn" text="Overview Not Found" />
					{/if}
				{:catch err}
					{#if err instanceof ApiError}
						<ErrorMessage variant="warn" text={err.message} retry={loadInvoiceOverview} />
					{:else}
						<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoiceOverview} />
					{/if}
				{/await}
			</div>
		</DashboardCard>

		<DashboardCard title="Invoices Paid" class="col-span-2 max-h-80 min-h-50">
			<div class="relative h-70 px-6 py-2">
				{#await invoiceMetricsPromise}
					<Spinner />
				{:then res}
					{#if res}
						<div class="absolute inset-x-6 inset-y-2">
							<Chart.Container config={invChartConfig} class="h-full w-full">
								<BarChart
									data={res}
									x="key"
									axis="x"
									yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
									seriesLayout="group"
									series={[
										{
											key: 'value',
											label: invChartConfig.paid.label,
											color: invChartConfig.paid.color
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
						<ErrorMessage variant="warn" text={err.message} retry={loadInvoiceMetrics} />
					{:else}
						<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoiceMetrics} />
					{/if}
				{/await}
			</div>
		</DashboardCard>
	</div>
</div>
