<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';

	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import * as Table from '$lib/components/ui/table';
	import type { Channel } from '../message/types';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import ChannelViewTicket from './ChannelViewTicket.svelte';

	interface Props {
		open: Boolean;
		close: () => void;
		channel?: Channel | null;
	}
	let { open = $bindable(), close, channel }: Props = $props();

	let page = $state(1);

	let tickets = [
		{
			subject: 'Unable to access billing portal',
			status: 'Resolved',
			createdOn: '2025-06-21T10:23:00Z',
			lastUpdated: '2025-06-22T09:12:00Z',
			createdBy: 'Alice Morgan'
		},
		{
			subject: 'Feature request: Dark mode for dashboard',
			status: 'Waiting For Reply',
			createdOn: '2025-06-19T14:55:00Z',
			lastUpdated: '2025-06-20T16:40:00Z',
			createdBy: 'John Taylor'
		},
		{
			subject: 'Error 502 when submitting form',
			status: 'In-Progress',
			createdOn: '2025-07-01T08:30:00Z',
			lastUpdated: '2025-07-03T11:22:00Z',
			createdBy: 'Carlos Hernandez'
		},
		{
			subject: 'Password reset not working',
			status: 'Resolved',
			createdOn: '2025-06-28T07:45:00Z',
			lastUpdated: '2025-06-28T08:10:00Z',
			createdBy: 'Carlos Hernandez'
		},
		{
			subject: 'App crashes on iOS 17',
			status: 'In-Progress',
			createdOn: '2025-07-05T12:10:00Z',
			lastUpdated: '2025-07-06T15:40:00Z',
			createdBy: 'Mina Kowalski'
		},
		{
			subject: 'Need invoice for May 2025',
			status: 'Resolved',
			createdOn: '2025-06-30T09:00:00Z',
			lastUpdated: '2025-06-30T09:15:00Z',
			createdBy: 'Derek Wilson'
		},
		{
			subject: 'How to integrate with Zapier?',
			status: 'Waiting For Reply',
			createdOn: '2025-07-02T13:42:00Z',
			lastUpdated: '2025-07-03T10:00:00Z',
			createdBy: 'Elena Petrova'
		},
		{
			subject: 'Two-factor auth setup not working',
			status: 'In-Progress',
			createdOn: '2025-06-27T18:25:00Z',
			lastUpdated: '2025-07-01T09:30:00Z',
			createdBy: 'James Liu'
		},
		{
			subject: 'Clarification on pricing tiers',
			status: 'Resolved',
			createdOn: '2025-07-01T10:15:00Z',
			lastUpdated: '2025-07-01T11:00:00Z',
			createdBy: 'Sophia Reyes'
		},
		{
			subject: 'Need to change account ownership',
			status: 'Waiting For Reply',
			createdOn: '2025-07-04T17:35:00Z',
			lastUpdated: '2025-07-05T08:12:00Z',
			createdBy: 'Ahmad Saleh'
		},
		{
			subject: 'Bug in PDF export function',
			status: 'In-Progress',
			createdOn: '2025-06-26T07:20:00Z',
			lastUpdated: '2025-06-30T13:00:00Z',
			createdBy: 'Julia Becker'
		},
		{
			subject: 'Account suspended after payment',
			status: 'Resolved',
			createdOn: '2025-06-25T16:45:00Z',
			lastUpdated: '2025-06-26T09:00:00Z',
			createdBy: 'Benjamin Tan'
		},
		{
			subject: 'Need help with API access token',
			status: 'Waiting For Reply',
			createdOn: '2025-07-03T10:05:00Z',
			lastUpdated: '2025-07-03T10:30:00Z',
			createdBy: 'Fatima Noor'
		},
		{
			subject: 'Unexpected charges in June bill',
			status: 'In-Progress',
			createdOn: '2025-07-06T14:20:00Z',
			lastUpdated: '2025-07-07T09:00:00Z',
			createdBy: 'Lucas Bennett'
		},
		{
			subject: 'Requesting data deletion',
			status: 'Resolved',
			createdOn: '2025-06-29T11:30:00Z',
			lastUpdated: '2025-06-29T12:00:00Z',
			createdBy: 'Naomi Tanaka'
		}
	];

	let selectedTicket = $state();

	let viewTicketDialog = createDialogState();
</script>

<FullScreenDialog bind:open {close}>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto flex items-end justify-between gap-4">
				<p class="font-bold">Support Tickets of {channel?.name}</p>
				<div class="w-full max-w-xs">
					<SearchBar />
				</div>
			</div>
		</div>

		<div class="overflow-y-auto">
			{#if tickets.length > 0}
				<Table.Root class="container mx-auto">
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Subject</Table.Head>
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
								<Table.Cell class="flex items-center gap-1">
									{entry.status}
									{#if entry.status == 'Resolved'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>
									{new Date(entry.createdOn).toLocaleString(undefined, {
										year: 'numeric',
										month: 'long',
										day: 'numeric',
										hour: 'numeric',
										minute: '2-digit',
										hour12: true
									})}
								</Table.Cell>
								<Table.Cell
									>{new Date(entry.lastUpdated).toLocaleString(undefined, {
										year: 'numeric',
										month: 'long',
										day: 'numeric',
										hour: 'numeric',
										minute: '2-digit',
										hour12: true
									})}</Table.Cell
								>
								<Table.Cell>{entry.createdBy}</Table.Cell>
								<Table.Cell>
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
									>
										<button
											title="View"
											onclick={() => {
												selectedTicket = entry;
												viewTicketDialog.open();
											}}
										>
											<ArrowRight size={16} />
										</button>
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</div>
		{#if tickets.length > 0}
			<div class="container mx-auto flex justify-end">
				<Pagination bind:page />
			</div>
		{/if}
	</div>
</FullScreenDialog>

<ChannelViewTicket
	bind:open={viewTicketDialog.isOpen}
	close={viewTicketDialog.close}
	ticket={selectedTicket}
/>
