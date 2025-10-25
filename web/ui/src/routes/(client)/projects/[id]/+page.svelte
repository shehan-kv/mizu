<script lang="ts">
	import { page } from '$app/state';
	import * as Table from '$lib/components/ui/table';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { getChangeRequestsByProject, type ChangeRequest } from '$lib/api/changeRequest';
	import { getContractsByProject, type Contract } from '$lib/api/contracts';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getFilesByProject, type File } from '$lib/api/files';
	import {
		getInvoicesByProjectId,
		getPaidInvoiceCountByProject,
		type InvoiceMetric,
		type InvoiceWithStatus
	} from '$lib/api/invoices';
	import {
		getProjectDetails,
		getTaskCompletedMetricsByProject,
		type ProjectDetails,
		type TaskMetric
	} from '$lib/api/projects';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import User from '$lib/components/User.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { onMount } from 'svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Checks from 'phosphor-svelte/lib/Checks';
	import KanbanTaskList from '$lib/components/KanbanTaskList.svelte';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import { BarChart, LineChart } from 'layerchart';
	import { curveLinear } from 'd3-shape';
	import { scalePoint } from 'd3-scale';

	let id = Number(page.params.id);

	let projectPromise: Promise<ProjectDetails> | null = $state(null);
	let projectAbort: AbortController | null = null;

	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}
		projectAbort = new AbortController();

		projectPromise = getProjectDetails(id, projectAbort.signal);
	}

	let invoicesPromise: Promise<PaginatedResponse<InvoiceWithStatus>> | null = $state(null);
	let invoicesAbort: AbortController | null = null;
	function loadInvoices() {
		if (invoicesAbort) {
			invoicesAbort.abort();
		}
		invoicesAbort = new AbortController();

		invoicesPromise = getInvoicesByProjectId(id, { page: 1, limit: 20 }, invoicesAbort.signal);
	}

	let contractsPromise: Promise<PaginatedResponse<Contract>> | null = $state(null);
	let contractsAbort: AbortController | null = null;
	function loadContracts() {
		if (contractsAbort) {
			contractsAbort.abort();
		}
		contractsAbort = new AbortController();

		contractsPromise = getContractsByProject(id, '', '', 1, 20, contractsAbort.signal);
	}

	let chReqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);
	let chReqAbort: AbortController | null = null;
	function loadChReqs() {
		if (chReqAbort) {
			chReqAbort.abort();
		}
		chReqAbort = new AbortController();

		chReqPromise = getChangeRequestsByProject(id, { page: 1, limit: 20 }, chReqAbort.signal);
	}

	let filesPromise: Promise<PaginatedResponse<File>> | null = $state(null);
	let filesAbort: AbortController | null = null;
	function loadFiles() {
		if (filesAbort) {
			filesAbort.abort();
		}
		filesAbort = new AbortController();

		filesPromise = getFilesByProject(id, '', 1, 20, filesAbort.signal);
	}

	let InvoicePaidCountPromise: Promise<InvoiceMetric[]> | null = $state(null);
	let InvoicePaidCountAbort: AbortController | null = null;
	function loadPaidInvoiceMetrics() {
		if (InvoicePaidCountAbort) {
			InvoicePaidCountAbort.abort();
		}
		InvoicePaidCountAbort = new AbortController();

		InvoicePaidCountPromise = getPaidInvoiceCountByProject(id, InvoicePaidCountAbort.signal);
	}

	let taskCompleteMetricsPromise: Promise<TaskMetric[]> | null = $state(null);
	let taskCompleteMetricsAbort: AbortController | null = null;
	function loadTaskCompleteMetrics() {
		if (taskCompleteMetricsAbort) {
			taskCompleteMetricsAbort.abort();
		}
		taskCompleteMetricsAbort = new AbortController();

		taskCompleteMetricsPromise = getTaskCompletedMetricsByProject(
			id,
			taskCompleteMetricsAbort.signal
		);
	}

	const invChartConfig = {
		paid: { label: 'Paid', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;

	const taskChartConfig = {
		completed: { label: 'Completed', color: 'var(--color-blue-400)' }
	} satisfies Chart.ChartConfig;

	onMount(() => {
		loadProject();
		loadPaidInvoiceMetrics();
		loadTaskCompleteMetrics();
		loadFiles();
		loadInvoices();
		loadContracts();
		loadChReqs();
	});
</script>

<svelte:head>
	<title>View Project</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<div class="space-y-2 rounded border p-6">
		{#await projectPromise}
			<Spinner />
		{:then res}
			{#if res}
				<div>
					<p class="text-xs text-neutral-500">Name</p>
					<p>{res.name}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Status</p>
					<div class="flex items-center gap-1.5 text-sm">
						{#if res.status == 'started'}
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
						{toTitleCase(res.status)}
					</div>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Created On</p>
					<p class="text-sm">{formatDate(res.createdAt)}</p>
				</div>
				<div class="mt-6 grid grid-cols-2 gap-x-20 gap-y-3">
					<div>
						<p class="text-xs text-neutral-500">Invoices</p>
						<p class="text-sm">
							{res.invoicePaidCount} / {res.invoiceCount} Paid
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Quotations</p>
						<p class="text-sm">{res.quoteCount}</p>
					</div>

					<div>
						<p class="text-xs text-neutral-500">Tasks</p>
						<p class="text-sm">
							{res.taskCompletedCount} / {res.taskCount} Completed
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Contracts</p>
						<p class="text-sm">
							{res.contractSignedCount} / {res.contractCount} Signed
						</p>
					</div>

					<div>
						<p class="text-xs text-neutral-500">Change Requests</p>
						<p class="text-sm">
							{res.changeReqClosedCount} / {res.changeReqCount} Closed
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Files</p>
						<p class="text-sm">{res.fileCount}</p>
					</div>
				</div>
			{:else}
				<ErrorMessage variant="warn" text="Project Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View This Project"
					retry={loadProject}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
			{/if}
		{/await}
	</div>

	<div
		class="min-h-50 max-h-100 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Members</p>
		</div>

		<div class="space-y-2 overflow-scroll px-6 py-4">
			{#await projectPromise}
				<Spinner />
			{:then res}
				{#if res && res.members.length > 0}
					{#each res.members as member}
						<User
							image={member.image}
							role={member.role}
							title={member.title}
							name={`${member.firstName} ${member.lastName}`}
						/>
					{/each}
				{:else}
					<ErrorMessage variant="warn" text="Members Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Members"
						retry={loadProject}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
				{/if}
			{/await}
		</div>
	</div>

	<div
		class="col-span-2 grid min-h-80 grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Invoices Paid</p>
		</div>
		<div class="px-6 py-2">
			{#await InvoicePaidCountPromise}
				<Spinner />
			{:then res}
				{#if res}
					<Chart.Container config={invChartConfig} class="h-70 w-full">
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
				{:else}
					<ErrorMessage variant="warn" text="Metrics Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Metrics"
						retry={loadProject}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
				{/if}
			{/await}
		</div>
	</div>

	<div class="h-84 col-span-4 grid grid-cols-4 gap-2 overflow-hidden">
		<div class="col-span-2 rounded border">
			<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
				<p class="text-sm">Tasks Completed</p>
			</div>
			<div class="px-6 py-2">
				{#await taskCompleteMetricsPromise}
					<Spinner />
				{:then res}
					{#if res && res.length > 0}
						<Chart.Container config={taskChartConfig} class="h-70 w-full">
							<BarChart
								data={res}
								x="key"
								axis="x"
								yDomain={[0, Math.max(1, ...res.map((d) => d.value))]}
								seriesLayout="group"
								series={[
									{
										key: 'value',
										label: taskChartConfig.completed.label,
										color: taskChartConfig.completed.color
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
					{#if err instanceof APIBadRequestError}
						<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
					{:else if err instanceof APIForbiddenError}
						<ErrorMessage
							variant="warn"
							text="You Don't Have Permission To View Metrics"
							retry={loadProject}
						/>
					{:else if err instanceof APINotFoundError}
						<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
					{:else if err instanceof APIServerError}
						<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
					{:else}
						<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
					{/if}
				{/await}
			</div>
		</div>

		<div
			class="col-span-2 grid grid-rows-[min-content_1fr]
			overflow-hidden rounded border"
		>
			<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
				<p class="text-sm">Files</p>
				<a href={`/projects/${id}/files`} class="flex items-center gap-1 text-sm">
					<span>View All</span>
					<ArrowRight />
				</a>
			</div>
			<div class="overflow-scroll px-6 py-2">
				{#await filesPromise}
					<Spinner />
				{:then res}
					{#if res && res.data.length > 0}
						<Table.Root>
							<Table.Body>
								{#each res.data as file}
									<Table.Row
										class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
									>
										<Table.Cell class="pl-0">
											{file.originalName}
										</Table.Cell>
										<Table.Cell>
											{file.user.firstName}
											{file.user.lastName}
										</Table.Cell>
										<Table.Cell class="pr-0" align="right">
											<a href={file.url} class="inline-block cursor-pointer">
												<DownloadSimple size={18} />
											</a>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{:else}
						<ErrorMessage variant="warn" text="Files Not Found" />
					{/if}
				{:catch err}
					{#if err instanceof APIBadRequestError}
						<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
					{:else if err instanceof APIForbiddenError}
						<ErrorMessage
							variant="warn"
							text="You Don't Have Permission To View Files"
							retry={loadProject}
						/>
					{:else if err instanceof APINotFoundError}
						<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
					{:else if err instanceof APIServerError}
						<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
					{:else}
						<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
					{/if}
				{/await}
			</div>
		</div>
	</div>

	<div
		class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Invoices / Quotes</p>
			<a href={`/projects/${id}/invoices`} class="flex items-center gap-1 text-sm">
				<span>View All</span>
				<ArrowRight />
			</a>
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
									<Table.Cell class="flex items-center gap-1">
										{toTitleCase(invoice.status)}
										{#if invoice.status == 'accepted'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</Table.Cell>
									<Table.Cell>
										{currencyFormatter(invoice.currencyCode, invoice.total)} Total
									</Table.Cell>
									<Table.Cell>
										Issued On {formatDate(invoice.issuedAt)}
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
					<ErrorMessage variant="warn" text="Invoices / Quotes Not Found" />
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Invoices/Quotes"
						retry={loadProject}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
				{/if}
			{/await}
		</div>
	</div>

	<div
		class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Contracts</p>
			<a href={`/projects/${id}/contracts`} class="flex items-center gap-1 text-sm">
				<span>View All</span>
				<ArrowRight />
			</a>
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
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Contracts"
						retry={loadProject}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
				{/if}
			{/await}
		</div>
	</div>

	<div
		class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
	>
		<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Change Requests</p>
			<a href={`/projects/${id}/change-requests`} class="flex items-center gap-1 text-sm">
				<span>View All</span>
				<ArrowRight />
			</a>
		</div>
		<div class="overflow-scroll px-6 py-2">
			{#await chReqPromise}
				<Spinner />
			{:then res}
				{#if res && res.data.length > 0}
					<Table.Root>
						<Table.Body>
							{#each res.data as chReq}
								<Table.Row
									class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
								>
									<Table.Cell class="pl-0">
										{chReq.title}
									</Table.Cell>
									<Table.Cell class="flex items-center gap-1">
										{toTitleCase(chReq.status)}
										{#if chReq.status == 'closed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</Table.Cell>
									<Table.Cell>
										Started By {chReq.requestedBy.firstName}
										{chReq.requestedBy.lastName}
									</Table.Cell>
									<Table.Cell>
										Created On {formatDate(chReq.createdAt)}
									</Table.Cell>
									<Table.Cell class="pr-0" align="right">
										<a href={`/change-requests/${chReq.id}`} title="View">
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
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View Change Requests"
						retry={loadProject}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
				{/if}
			{/await}
		</div>
	</div>

	<div class="max-h-100 col-span-4 grid grid-rows-[min-content_1fr] overflow-hidden rounded border">
		<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Kanban Board</p>
			<a href={`/projects/${id}/kanban`} class="flex items-center gap-1 text-sm">
				<span>View</span>
				<ArrowRight />
			</a>
		</div>
		<div class="grid grid-cols-3 gap-2 overflow-scroll px-6 py-2">
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Backlog</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="backlog" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">In-Progress</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="in-progress" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Completed</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="completed" />
				</div>
			</div>
		</div>
	</div>
</div>
