<script lang="ts">
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Envelope from 'phosphor-svelte/lib/Envelope';

	import * as Table from '$lib/components/ui/table';
	import type { Channel } from '../message/types';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { onMount } from 'svelte';
	import Spinner from '../Spinner.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import ChannelViewContract from './ChannelViewContract.svelte';

	interface Props {
		open: Boolean;
		close: () => void;
		channel?: Channel | null;
	}
	let { open = $bindable(), close, channel }: Props = $props();

	let isLoading = $state(false);
	let page = $state(1);

	interface Contract {
		id: number;
		name: string;
		status: string;
		date: string;
		lastUpdated: string;
		parties: {
			name: string;
			signed: boolean;
		}[];
	}

	let contracts: Contract[] = [
		{
			id: 1,
			name: 'Website Redesign',
			status: 'Signed',
			date: '2025-05-01',
			lastUpdated: '2025-05-10T12:00:00Z',
			parties: [
				{ name: 'Alice Freelancer', signed: true },
				{ name: 'Bob Client', signed: true },
				{ name: 'Yvonne Freelancer', signed: true },
				{ name: 'Zack Client', signed: false }
			]
		},
		{
			id: 2,
			name: 'Mobile App Development',
			status: 'Pending',
			date: '2025-05-15',
			lastUpdated: '2025-05-20T09:30:00Z',
			parties: [
				{ name: 'Charlie Freelancer', signed: false },
				{ name: 'Dana Client', signed: false }
			]
		},
		{
			id: 3,
			name: 'SEO Optimization',
			status: 'Signed',
			date: '2025-04-20',
			lastUpdated: '2025-04-25T16:45:00Z',
			parties: [
				{ name: 'Eve Freelancer', signed: true },
				{ name: 'Frank Client', signed: true }
			]
		},
		{
			id: 4,
			name: 'Content Writing',
			status: 'Rejected',
			date: '2025-03-10',
			lastUpdated: '2025-03-15T11:00:00Z',
			parties: [
				{ name: 'Grace Freelancer', signed: false },
				{ name: 'Hank Client', signed: false }
			]
		},
		{
			id: 4,
			name: 'Logo Design',
			status: 'Signed',
			date: '2025-06-01',
			lastUpdated: '2025-06-05T14:20:00Z',
			parties: [
				{ name: 'Ivy Freelancer', signed: true },
				{ name: 'Jack Client', signed: true }
			]
		},
		{
			id: 5,
			name: 'Social Media Management',
			status: 'Pending',
			date: '2025-06-10',
			lastUpdated: '2025-06-12T08:15:00Z',
			parties: [
				{ name: 'Kate Freelancer', signed: false },
				{ name: 'Leo Client', signed: false }
			]
		},
		{
			id: 6,
			name: 'Video Production',
			status: 'Signed',
			date: '2025-05-22',
			lastUpdated: '2025-05-28T10:00:00Z',
			parties: [
				{ name: 'Mia Freelancer', signed: true },
				{ name: 'Nate Client', signed: true }
			]
		},
		{
			id: 7,
			name: 'Data Analysis',
			status: 'Signed',
			date: '2025-04-30',
			lastUpdated: '2025-05-02T13:30:00Z',
			parties: [
				{ name: 'Olivia Freelancer', signed: true },
				{ name: 'Paul Client', signed: true }
			]
		},
		{
			id: 8,
			name: 'Customer Support',
			status: 'Pending',
			date: '2025-06-15',
			lastUpdated: '2025-06-18T09:00:00Z',
			parties: [
				{ name: 'Quinn Freelancer', signed: false },
				{ name: 'Rachel Client', signed: false }
			]
		},
		{
			id: 9,
			name: 'Marketing Strategy',
			status: 'Signed',
			date: '2025-05-05',
			lastUpdated: '2025-05-10T15:45:00Z',
			parties: [
				{ name: 'Sam Freelancer', signed: true },
				{ name: 'Tina Client', signed: true }
			]
		},
		{
			id: 10,
			name: 'App Testing',
			status: 'Rejected',
			date: '2025-03-25',
			lastUpdated: '2025-03-30T12:30:00Z',
			parties: [
				{ name: 'Uma Freelancer', signed: false },
				{ name: 'Victor Client', signed: false }
			]
		},
		{
			id: 11,
			name: 'Graphic Design',
			status: 'Signed',
			date: '2025-06-03',
			lastUpdated: '2025-06-07T11:15:00Z',
			parties: [
				{ name: 'Wendy Freelancer', signed: true },
				{ name: 'Xander Client', signed: true }
			]
		},
		{
			id: 12,
			name: 'Translation Services',
			status: 'Pending',
			date: '2025-06-20',
			lastUpdated: '2025-06-21T10:00:00Z',
			parties: [
				{ name: 'Yara Freelancer', signed: false },
				{ name: 'Zane Client', signed: false }
			]
		},
		{
			id: 13,
			name: 'IT Support',
			status: 'Signed',
			date: '2025-05-18',
			lastUpdated: '2025-05-22T14:00:00Z',
			parties: [
				{ name: 'Aaron Freelancer', signed: true },
				{ name: 'Beth Client', signed: true }
			]
		},
		{
			id: 14,
			name: 'Legal Consulting',
			status: 'Signed',
			date: '2025-04-10',
			lastUpdated: '2025-04-15T09:30:00Z',
			parties: [
				{ name: 'Carl Freelancer', signed: true },
				{ name: 'Diana Client', signed: true }
			]
		},
		{
			id: 15,
			name: 'Project Management',
			status: 'Pending',
			date: '2025-06-12',
			lastUpdated: '2025-06-14T10:45:00Z',
			parties: [
				{ name: 'Ethan Freelancer', signed: false },
				{ name: 'Fiona Client', signed: false }
			]
		},
		{
			id: 16,
			name: 'Copywriting',
			status: 'Signed',
			date: '2025-05-25',
			lastUpdated: '2025-05-28T16:00:00Z',
			parties: [
				{ name: 'Gina Freelancer', signed: true },
				{ name: 'Harry Client', signed: true }
			]
		},
		{
			id: 17,
			name: 'HR Consulting',
			status: 'Rejected',
			date: '2025-03-05',
			lastUpdated: '2025-03-10T13:00:00Z',
			parties: [
				{ name: 'Irene Freelancer', signed: false },
				{ name: 'Jason Client', signed: false }
			]
		},
		{
			id: 18,
			name: 'Financial Auditing',
			status: 'Signed',
			date: '2025-06-08',
			lastUpdated: '2025-06-11T11:30:00Z',
			parties: [
				{ name: 'Karen Freelancer', signed: true },
				{ name: 'Liam Client', signed: true }
			]
		},
		{
			id: 19,
			name: 'Product Design',
			status: 'Pending',
			date: '2025-06-16',
			lastUpdated: '2025-06-19T09:45:00Z',
			parties: [
				{ name: 'Mason Freelancer', signed: false },
				{ name: 'Nina Client', signed: false }
			]
		},
		{
			id: 20,
			name: 'Advertising Campaign',
			status: 'Signed',
			date: '2025-05-12',
			lastUpdated: '2025-05-17T14:30:00Z',
			parties: [
				{ name: 'Oscar Freelancer', signed: true },
				{ name: 'Paula Client', signed: true }
			]
		},
		{
			id: 21,
			name: 'Event Planning',
			status: 'Signed',
			date: '2025-04-28',
			lastUpdated: '2025-05-01T10:15:00Z',
			parties: [
				{ name: 'Quincy Freelancer', signed: true },
				{ name: 'Rita Client', signed: true }
			]
		},
		{
			id: 22,
			name: 'Photography',
			status: 'Pending',
			date: '2025-06-14',
			lastUpdated: '2025-06-16T08:50:00Z',
			parties: [
				{ name: 'Steve Freelancer', signed: false },
				{ name: 'Tracy Client', signed: false }
			]
		},
		{
			id: 23,
			name: 'UX/UI Design',
			status: 'Signed',
			date: '2025-05-08',
			lastUpdated: '2025-05-13T12:20:00Z',
			parties: [
				{ name: 'Uma Freelancer', signed: true },
				{ name: 'Vince Client', signed: true }
			]
		},
		{
			id: 24,
			name: 'Consulting Services',
			status: 'Rejected',
			date: '2025-03-18',
			lastUpdated: '2025-03-22T14:10:00Z',
			parties: [
				{ name: 'Walt Freelancer', signed: false },
				{ name: 'Xena Client', signed: false }
			]
		},
		{
			id: 25,
			name: 'Technical Writing',
			status: 'Signed',
			date: '2025-06-06',
			lastUpdated: '2025-06-09T15:00:00Z',
			parties: [
				{ name: 'Yvonne Freelancer', signed: true },
				{ name: 'Zack Client', signed: true }
			]
		},
		{
			id: 26,
			name: 'Research Project',
			status: 'Pending',
			date: '2025-06-21',
			lastUpdated: '2025-06-22T10:30:00Z',
			parties: [
				{ name: 'Aaron Freelancer', signed: false },
				{ name: 'Beth Client', signed: false }
			]
		},
		{
			id: 27,
			name: 'Training Program',
			status: 'Signed',
			date: '2025-05-30',
			lastUpdated: '2025-06-02T09:40:00Z',
			parties: [
				{ name: 'Carl Freelancer', signed: true },
				{ name: 'Diana Client', signed: true }
			]
		},
		{
			id: 28,
			name: 'Software Maintenance',
			status: 'Signed',
			date: '2025-04-15',
			lastUpdated: '2025-04-20T11:55:00Z',
			parties: [
				{ name: 'Ethan Freelancer', signed: true },
				{ name: 'Fiona Client', signed: true }
			]
		},
		{
			id: 29,
			name: 'Brand Strategy',
			status: 'Pending',
			date: '2025-06-17',
			lastUpdated: '2025-06-20T14:25:00Z',
			parties: [
				{ name: 'Gina Freelancer', signed: false },
				{ name: 'Harry Client', signed: false }
			]
		}
	];

	function getPartyNamesText(parties: { name: string; signed: boolean }[]) {
		if (parties.length == 1) return { omitted: 0, names: parties[0].name };

		if (parties.length > 1) {
			return {
				omitted: parties.length - 2,
				names: `${parties[0].name.split(' ')[0]} / ${parties[1].name.split(' ')[0]}`
			};
		}

		return { omitted: 0, names: 'N/A' };
	}

	let selectedContract: Contract | null = $state(contracts[0]);
	let contractViewDialog = createDialogState();

	$effect(() => {
		let timeout: number;

		if (open && channel) {
			// TODO: Added to simulate a network call. Remove
			isLoading = true;
			timeout = setTimeout(() => {
				isLoading = false;
			}, 2000);
		}

		// TODO: Fetch contracts here

		return () => {
			// TODO: Remove clearTimeout
			clearTimeout(timeout);
		};
	});
</script>

{#snippet parties(props: { omitted: number; names: string })}
	<p>{props.names}</p>
	{#if props.omitted > 0}
		<p class="mt-0.5 text-xs text-neutral-500 dark:text-neutral-400">
			+{props.omitted}
			{props.omitted == 1 ? 'Other' : 'Others'}
		</p>
	{/if}
{/snippet}

<FullScreenDialog bind:open {close}>
	{#if isLoading}
		<Spinner />
	{:else if channel}
		<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
			<div class="flex-none">
				<div class="container mx-auto flex items-end justify-between gap-4">
					<p class="font-bold">Contracts of {channel.name}</p>
					<div class="w-full max-w-xs">
						<SearchBar />
					</div>
				</div>
			</div>

			<div class="overflow-y-auto">
				{#if contracts.length > 0}
					<Table.Root class="container mx-auto">
						<Table.Header>
							<Table.Row>
								<Table.Head class="font-bold">Name</Table.Head>
								<Table.Head class="font-bold">Status</Table.Head>
								<Table.Head class="font-bold">Parties</Table.Head>
								<Table.Head class="font-bold">Date</Table.Head>
								<Table.Head class="font-bold">Last Updated</Table.Head>
								<Table.Head class="font-bold">Actions</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each contracts as contract}
								<Table.Row>
									<Table.Cell>{contract.name}</Table.Cell>
									<Table.Cell>{contract.status}</Table.Cell>
									<Table.Cell>
										<div>
											{@render parties(getPartyNamesText(contract.parties))}
										</div>
									</Table.Cell>
									<Table.Cell>{contract.date}</Table.Cell>
									<Table.Cell>{contract.lastUpdated}</Table.Cell>
									<Table.Cell>
										<div
											class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
										>
											<button
												onclick={() => {
													selectedContract = contract;
													contractViewDialog.open();
												}}
												title="View Contract"
											>
												<ArrowRight size={18} />
											</button>
											<button title="Download the Latest Version as PDF">
												<DownloadSimple size={18} />
											</button>
											<button title="Email Me"><Envelope size={18} /></button>
										</div>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{/if}
			</div>
			{#if contracts.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page />
				</div>
			{/if}
		</div>
	{:else}
		<ErrorMessage variant="info" text="Channel Not Selected" />
	{/if}
</FullScreenDialog>

<ChannelViewContract
	bind:open={contractViewDialog.isOpen}
	close={contractViewDialog.close}
	contract={selectedContract}
/>
