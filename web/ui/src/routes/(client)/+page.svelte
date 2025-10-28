<script lang="ts">
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import {
		getProjectCreatedMetrics,
		getProjects,
		type Project,
		type ProjectMetric
	} from '$lib/api/projects';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import * as Table from '$lib/components/ui/table';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { scalePoint } from 'd3-scale';
	import { curveLinear } from 'd3-shape';
	import { LineChart } from 'layerchart';
	import { onMount } from 'svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { formatDate } from '$lib/utils/formatDate';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { getContracts, type Contract } from '$lib/api/contracts';
	import {
		getInvoiceOverview,
		getInvoices,
		getPaidInvoiceMetrics,
		type InvoiceMetric,
		type InvoiceOverview,
		type InvoiceOverviewMetric,
		type InvoiceSummary
	} from '$lib/api/invoices';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDecimalSuffix } from '$lib/utils/formatDecimalSuffix';
	import { getChangeRequests, type ChangeRequest } from '$lib/api/changeRequest';

	let projectsPromise: Promise<PaginatedResponse<Project>> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProjects() {
		if (projectAbort) {
			projectAbort.abort();
		}
		projectAbort = new AbortController();

		projectsPromise = getProjects('', 1, 20, '', projectAbort.signal);
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

		contractsPromise = getContracts('', '', 1, 20, contractsAbort.signal);
	}

	let invoicesPromise: Promise<PaginatedResponse<InvoiceSummary>> | null = $state(null);
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

		invoiceMetricsPromise = getPaidInvoiceMetrics(invoiceMetricsAbort.signal);
	}

	let invoiceOverviewPromise: Promise<InvoiceOverview> | null = $state(null);
	let invoiceOverviewAbort: AbortController | null = null;
	function loadInvoiceOverview() {
		if (invoiceOverviewAbort) {
			invoiceOverviewAbort.abort();
		}

		invoiceOverviewAbort = new AbortController();

		invoiceOverviewPromise = getInvoiceOverview(invoiceOverviewAbort.signal);
	}

	let chReqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);
	let chReqAbort: AbortController | null = null;
	function loadChReq() {
		if (chReqAbort) {
			chReqAbort.abort();
		}

		chReqAbort = new AbortController();

		chReqPromise = getChangeRequests('', '', 1, 20, chReqAbort.signal);
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
		loadChReq();
	});
</script>

<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<div
		class="col-span-2 grid min-h-80 grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Projects Created</p>
		</div>
		<div class="relative px-6 py-2">
			{#await projectMetricsPromise}
				<Spinner />
			{:then res}
				{#if res}
					<div class="absolute inset-x-6 inset-y-2">
						<Chart.Container config={invChartConfig} class="h-full w-full">
							<LineChart
								data={res}
								x="key"
								xScale={scalePoint()}
								yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
								axis="x"
								series={[
									{
										key: 'value',
										label: invChartConfig.paid.label,
										color: invChartConfig.paid.color
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
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProjectCountMetrics} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Metrics"
						retry={loadProjectCountMetrics}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProjectCountMetrics} />
				{:else if err instanceof APIServerError}
					<ErrorMessage
						variant="warn"
						text="Server Ran Into An Error"
						retry={loadProjectCountMetrics}
					/>
				{:else}
					<ErrorMessage
						variant="warn"
						text="An Unexpected Error Occured"
						retry={loadProjectCountMetrics}
					/>
				{/if}
			{/await}
		</div>
	</div>
	<div
		class="col-span-2 grid min-h-80 grid-rows-[min-content_1fr]
			overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Recent Projects</p>
		</div>
		<div class="overflow-scroll px-6 py-2">
			{#await projectsPromise}
				<Spinner />
			{:then res}
				{#if res && res.data.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.data as project}
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
										<a href={`/projects/${project.id}`} title="View">
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
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProjects} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Metrics"
						retry={loadProjects}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProjects} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProjects} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProjects} />
				{/if}
			{/await}
		</div>
	</div>

	<div
		class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Recent Contracts</p>
		</div>
		<div class="overflow-scroll px-6 py-2">
			{#await contractsPromise}
				<Spinner />
			{:then res}
				{#if res && res.data.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.data as contract}
								<Table.Row
									class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
								>
									<Table.Cell class="pl-0">
										{contract.name}
									</Table.Cell>
									<Table.Cell>
										{contract.projectName}
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
									<Table.Cell>
										{contract.versions}
										{contract.versions == 1 ? 'Version' : 'Versions'}
									</Table.Cell>
									<Table.Cell class="pr-0" align="right">
										<a href={`/contracts/${contract.id}`} title="View">
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
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadContracts} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Contracts"
						retry={loadContracts}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadContracts} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadContracts} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadContracts} />
				{/if}
			{/await}
		</div>
	</div>

	<div class="col-span-4 grid grid-cols-5 gap-2">
		<div
			class="min-h-50 col-span-2 grid max-h-80 grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
		>
			<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
				<p class="text-sm">Recent Invoices</p>
			</div>
			<div class="overflow-scroll px-6 py-2">
				{#await invoicesPromise}
					<Spinner />
				{:then res}
					{#if res && res.data.length > 0}
						<Table.Root>
							<Table.Body>
								{#each res.data as invoice}
									<Table.Row
										class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
									>
										<Table.Cell class="pl-0">
											{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id}
										</Table.Cell>
										<Table.Cell align="right">
											{currencyFormatter(invoice.currencyCode, invoice.total)} Total
										</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(invoice.status)}
											{#if invoice.status == 'paid' || invoice.status == 'accepted'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell class="pr-0" align="right">
											<a href={`/invoices-and-quotes/${invoice.id}`} title="View">
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
					{#if err instanceof APIBadRequestError}
						<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoices} />
					{:else if err instanceof APIForbiddenError}
						<ErrorMessage
							variant="warn"
							text="You Don't Have Permission To View Invoices/Quotes"
							retry={loadInvoices}
						/>
					{:else if err instanceof APINotFoundError}
						<ErrorMessage variant="info" text="Not Found" retry={loadInvoices} />
					{:else if err instanceof APIServerError}
						<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadInvoices} />
					{:else}
						<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadInvoices} />
					{/if}
				{/await}
			</div>
		</div>
		<div
			class="min-h-50 grid max-h-80 grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
		>
			<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
				<p class="text-sm">Invoices Overview</p>
			</div>
			<div class="space-y-4 overflow-scroll px-6 py-2">
				{#snippet overview(label: string, items: InvoiceOverviewMetric[], isInvoice?: boolean)}
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
						{@render overview('Invoices Paid', res.paid, true)}
						{@render overview('Invoices Pending', res.pending, true)}
						{@render overview('Invoices Accepted', res.accepted, true)}
						{@render overview('Invoices Rejected', res.rejected, true)}
						{@render overview('Invoices Cancelled', res.cancelled, true)}
						{@render overview('Quotes Pending', res.quotesPending)}
						{@render overview('Quotes Rejected', res.quotesRejected)}
					{:else}
						<ErrorMessage variant="warn" text="Overview Not Found" />
					{/if}
				{:catch err}
					{#if err instanceof APIBadRequestError}
						<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoiceOverview} />
					{:else if err instanceof APIForbiddenError}
						<ErrorMessage
							variant="warn"
							text="You Don't Have Permission To View The Overview"
							retry={loadInvoiceOverview}
						/>
					{:else if err instanceof APINotFoundError}
						<ErrorMessage variant="info" text="Not Found" retry={loadInvoiceOverview} />
					{:else if err instanceof APIServerError}
						<ErrorMessage
							variant="warn"
							text="Server Ran Into An Error"
							retry={loadInvoiceOverview}
						/>
					{:else}
						<ErrorMessage
							variant="warn"
							text="An Unexpected Error Occured"
							retry={loadInvoiceOverview}
						/>
					{/if}
				{/await}
			</div>
		</div>
		<div
			class="col-span-2 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
		>
			<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
				<p class="text-sm">Invoices Paid</p>
			</div>
			<div class="h-70 relative px-6 py-2">
				{#await invoiceMetricsPromise}
					<Spinner />
				{:then res}
					{#if res}
						<div class="absolute inset-x-6 inset-y-2">
							<Chart.Container config={invChartConfig} class="h-full w-full">
								<LineChart
									data={res}
									x="key"
									xScale={scalePoint()}
									yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
									axis="x"
									series={[
										{
											key: 'value',
											label: invChartConfig.paid.label,
											color: invChartConfig.paid.color
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
						<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoiceMetrics} />
					{:else if err instanceof APIForbiddenError}
						<ErrorMessage
							variant="warn"
							text="You Don't Have Permission To View Metrics"
							retry={loadInvoiceMetrics}
						/>
					{:else if err instanceof APINotFoundError}
						<ErrorMessage variant="info" text="Not Found" retry={loadInvoiceMetrics} />
					{:else if err instanceof APIServerError}
						<ErrorMessage
							variant="warn"
							text="Server Ran Into An Error"
							retry={loadInvoiceMetrics}
						/>
					{:else}
						<ErrorMessage
							variant="warn"
							text="An Unexpected Error Occured"
							retry={loadInvoiceMetrics}
						/>
					{/if}
				{/await}
			</div>
		</div>
	</div>

	<div
		class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Recent Change Requests</p>
		</div>
		<div class="overflow-scroll px-6 py-2">
			{#await chReqPromise}
				<Spinner />
			{:then res}
				{#if res && res.data.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.data as req}
								<Table.Row
									class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
								>
									<Table.Cell class="pl-0">
										{req.title}
									</Table.Cell>
									<Table.Cell>
										{req.project.name}
									</Table.Cell>
									<Table.Cell class="flex items-center gap-1">
										{toTitleCase(req.status)}
										{#if req.status == 'closed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</Table.Cell>
									<Table.Cell>
										Created On {formatDate(req.createdAt)}
									</Table.Cell>
									<Table.Cell>
										Started By {req.requestedBy.firstName}
										{req.requestedBy.lastName}
									</Table.Cell>
									<Table.Cell class="pr-0" align="right">
										<a href={`/change-requests/${req.id}`} title="View">
											<ArrowRight size={18} />
										</a>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{:else}
					<ErrorMessage variant="warn" text="Change Requests Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadChReq} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Change Requests"
						retry={loadChReq}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadChReq} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadChReq} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadChReq} />
				{/if}
			{/await}
		</div>
	</div>
</div>
