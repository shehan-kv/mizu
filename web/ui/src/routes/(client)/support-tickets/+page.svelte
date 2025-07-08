<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { tick } from 'svelte';

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

	let page = $state(1);
</script>

<svelte:head>
	<title>Invoices and Quotes</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto lg:container">
		<div class="max-w-96">
			<SearchBar />
		</div>
	</div>

	<div class="mx-auto gap-4 overflow-y-auto lg:container">
		<div class="overflow-y-auto">
			{#if tickets.length > 0}
				<Table.Root class="container mx-auto">
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Subject</Table.Head>
							<Table.Head class="font-bold">Project</Table.Head>
							<Table.Head class="font-bold">Status</Table.Head>
							<Table.Head class="font-bold">Created On</Table.Head>
							<Table.Head class="font-bold">Last Updated</Table.Head>
							<Table.Head class="font-bold">Created By</Table.Head>
							<Table.Head class="font-bold">Actions</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each tickets as entry}
							<Table.Row>
								<Table.Cell>{entry.subject}</Table.Cell>
								<Table.Cell>{entry.projectName}</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{entry.status}
									{#if entry.status == 'Resolved'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>{entry.createdOn}</Table.Cell>
								<Table.Cell>{entry.lastUpdated}</Table.Cell>
								<Table.Cell>{entry.createdBy}</Table.Cell>
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
			<Pagination bind:page />
		</div>
	{/if}
</div>
