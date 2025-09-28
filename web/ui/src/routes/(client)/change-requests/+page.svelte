<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getChangeRequests, type ChangeRequest } from '$lib/api/changeRequest';
	import Spinner from '$lib/components/Spinner.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';

	let tickets = [
		{
			subject: 'Unable to access billing portal',
			status: 'Resolved',
			createdOn: '2025-06-21T10:23:00Z',
			lastUpdated: '2025-06-22T09:12:00Z',
			createdBy: 'Alice Morgan',
			projectName: 'Billing System Revamp'
		},
		{
			subject: 'Feature request: Dark mode for dashboard',
			status: 'Waiting For Reply',
			createdOn: '2025-06-19T14:55:00Z',
			lastUpdated: '2025-06-20T16:40:00Z',
			createdBy: 'John Taylor',
			projectName: 'UX/UI Enhancements'
		},
		{
			subject: 'Error 502 when submitting form',
			status: 'In-Progress',
			createdOn: '2025-07-01T08:30:00Z',
			lastUpdated: '2025-07-03T11:22:00Z',
			createdBy: 'Carlos Hernandez',
			projectName: 'Form Submission Stability'
		},
		{
			subject: 'Password reset not working',
			status: 'Resolved',
			createdOn: '2025-06-28T07:45:00Z',
			lastUpdated: '2025-06-28T08:10:00Z',
			createdBy: 'Carlos Hernandez',
			projectName: 'Authentication Improvements'
		},
		{
			subject: 'App crashes on iOS 17',
			status: 'In-Progress',
			createdOn: '2025-07-05T12:10:00Z',
			lastUpdated: '2025-07-06T15:40:00Z',
			createdBy: 'Mina Kowalski',
			projectName: 'Mobile App iOS Compatibility'
		},
		{
			subject: 'Need invoice for May 2025',
			status: 'Resolved',
			createdOn: '2025-06-30T09:00:00Z',
			lastUpdated: '2025-06-30T09:15:00Z',
			createdBy: 'Derek Wilson',
			projectName: 'Billing System Revamp'
		},
		{
			subject: 'How to integrate with Zapier?',
			status: 'Waiting For Reply',
			createdOn: '2025-07-02T13:42:00Z',
			lastUpdated: '2025-07-03T10:00:00Z',
			createdBy: 'Elena Petrova',
			projectName: 'Third-Party Integrations'
		},
		{
			subject: 'Two-factor auth setup not working',
			status: 'In-Progress',
			createdOn: '2025-06-27T18:25:00Z',
			lastUpdated: '2025-07-01T09:30:00Z',
			createdBy: 'James Liu',
			projectName: 'Authentication Improvements'
		},
		{
			subject: 'Clarification on pricing tiers',
			status: 'Resolved',
			createdOn: '2025-07-01T10:15:00Z',
			lastUpdated: '2025-07-01T11:00:00Z',
			createdBy: 'Sophia Reyes',
			projectName: 'Pricing Strategy Update'
		},
		{
			subject: 'Need to change account ownership',
			status: 'Waiting For Reply',
			createdOn: '2025-07-04T17:35:00Z',
			lastUpdated: '2025-07-05T08:12:00Z',
			createdBy: 'Ahmad Saleh',
			projectName: 'Account Management Enhancements'
		},
		{
			subject: 'Bug in PDF export function',
			status: 'In-Progress',
			createdOn: '2025-06-26T07:20:00Z',
			lastUpdated: '2025-06-30T13:00:00Z',
			createdBy: 'Julia Becker',
			projectName: 'Reporting Module Fixes'
		},
		{
			subject: 'Account suspended after payment',
			status: 'Resolved',
			createdOn: '2025-06-25T16:45:00Z',
			lastUpdated: '2025-06-26T09:00:00Z',
			createdBy: 'Benjamin Tan',
			projectName: 'Billing System Revamp'
		},
		{
			subject: 'Need help with API access token',
			status: 'Waiting For Reply',
			createdOn: '2025-07-03T10:05:00Z',
			lastUpdated: '2025-07-03T10:30:00Z',
			createdBy: 'Fatima Noor',
			projectName: 'API Development'
		},
		{
			subject: 'Unexpected charges in June bill',
			status: 'In-Progress',
			createdOn: '2025-07-06T14:20:00Z',
			lastUpdated: '2025-07-07T09:00:00Z',
			createdBy: 'Lucas Bennett',
			projectName: 'Billing System Revamp'
		},
		{
			subject: 'Requesting data deletion',
			status: 'Resolved',
			createdOn: '2025-06-29T11:30:00Z',
			lastUpdated: '2025-06-29T12:00:00Z',
			createdBy: 'Naomi Tanaka',
			projectName: 'Data Privacy Compliance'
		}
	];

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let type = $state(params.get('type') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let chReqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadInvoices() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		chReqPromise = getChangeRequests(q, status, pageNum, limit, abortController.signal);
	}

	function updateUrlParam() {
		if (q) {
			params.set('q', q);
		} else {
			params.delete('q');
		}

		params.set('page', pageNum.toString());
		params.set('limit', limit.toString());

		if (status) {
			params.set('status', status);
		} else {
			params.delete('status');
		}

		if (type) {
			params.set('type', type);
		} else {
			params.delete('type');
		}

		history.replaceState(null, '', `?${params.toString()}`);
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadInvoices();
	}

	onMount(() => {
		loadInvoices();
	});
</script>

<svelte:head>
	<title>Change Requests</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto flex gap-4 lg:container">
		<div class="max-w-96">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
		<div class="flex gap-2">
			<FilterSelect
				bind:value={status}
				onchange={handleFilter}
				name="Status"
				options={[
					{ value: '', label: 'All' },
					{ value: 'in-progress', label: 'In Progress' },
					{ value: 'waiting', label: 'Waiting' },
					{ value: 'closed', label: 'Closed' }
				]}
			/>
			<FilterInput
				id="limit"
				max={MAX_LIMIT}
				min={MIN_LIMIT}
				label="Limit"
				type="number"
				bind:value={limit}
				onchange={handleFilter}
			/>
		</div>
	</div>

	{#await chReqPromise}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Change Requests Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Subject</Table.Head>
									<Table.Head class="font-bold">Project</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Created On</Table.Head>
									<Table.Head class="font-bold">Created By</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as req}
									<Table.Row>
										<Table.Cell>{req.title}</Table.Cell>
										<Table.Cell>{req.project.name}</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(req.status)}
											{#if req.status == 'closed'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell>{new Date(req.createdAt).toLocaleString()}</Table.Cell>
										<Table.Cell>{req.requestedBy.firstName} {req.requestedBy.lastName}</Table.Cell>
										<Table.Cell>
											<button
												class="cursor-pointer px-1.5 text-xs
										text-neutral-500 hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-neutral-50"
												title="View"
											>
												<ArrowRight size={18} />
											</button>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</div>
			{#if tickets.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={30} perPage={limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof APIBadRequestError}
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoices} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Change Requests"
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
