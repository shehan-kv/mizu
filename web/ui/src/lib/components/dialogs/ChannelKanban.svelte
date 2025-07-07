<script lang="ts">
	import CellSignalMedium from 'phosphor-svelte/lib/CellSignalMedium';
	import CellSignalLow from 'phosphor-svelte/lib/CellSignalLow';
	import CellSignalFull from 'phosphor-svelte/lib/CellSignalFull';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import type { Channel } from '$lib/components/message/types';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';

	interface Props {
		open: Boolean;
		close: () => void;
		channel?: Channel | null;
	}
	let { open = $bindable(), close, channel }: Props = $props();

	let isLoading = $state(false);
	let page = $state(1);

	const files = [
		{
			fileName: 'Welcome_Guide_For_Clients.pdf',
			uploadedDate: '2025-06-14T09:50:00Z',
			uploadedBy: 'Olivia Smith',
			link: '/',
			size: '1.2 MB'
		},
		{
			fileName: 'freelancer_community_rules.docx',
			uploadedDate: '2025-06-14T08:15:00Z',
			uploadedBy: 'Liam Johnson',
			link: '/',
			size: '350 KB'
		},
		{
			fileName: 'General_Pricing_Sheet.xlsx',
			uploadedDate: '2025-06-13T17:20:00Z',
			uploadedBy: 'Ava Davis',
			link: '/',
			size: '450 KB'
		},
		{
			fileName: 'platform_update_notes.txt',
			uploadedDate: '2025-06-13T11:45:00Z',
			uploadedBy: 'Noah Garcia',
			link: '/',
			size: '15 KB'
		},
		{
			fileName: 'Service_Agreement_Template.pdf',
			uploadedDate: '2025-06-12T09:30:00Z',
			uploadedBy: 'Sophia Rodriguez',
			link: '/',
			size: '900 KB'
		},
		{
			fileName: 'sample_portfolio_items.zip',
			uploadedDate: '2025-06-11T14:00:00Z',
			uploadedBy: 'Jackson Martinez',
			link: '/',
			size: '5.4 MB'
		},
		{
			fileName: 'How_To_Submit_Feedback.pptx',
			uploadedDate: '2025-06-11T09:10:00Z',
			uploadedBy: 'Olivia Smith',
			link: '/',
			size: '2.1 MB'
		},
		{
			fileName: 'common_issues_troubleshoot.docx',
			uploadedDate: '2025-06-10T16:55:00Z',
			uploadedBy: 'Liam Johnson',
			link: '/',
			size: '400 KB'
		},
		{
			fileName: 'Preferred_Communication_Methods.pdf',
			uploadedDate: '2025-06-09T10:25:00Z',
			uploadedBy: 'Ava Davis',
			link: '/',
			size: '1.0 MB'
		},
		{
			fileName: 'terms_of_use_v2.pdf',
			uploadedDate: '2025-06-08T15:00:00Z',
			uploadedBy: 'Noah Garcia',
			link: '/',
			size: '850 KB'
		},
		{
			fileName: 'Monthly_Platform_Report.xlsx',
			uploadedDate: '2025-06-07T12:00:00Z',
			uploadedBy: 'Sophia Rodriguez',
			link: '/',
			size: '600 KB'
		},
		{
			fileName: 'previous_chat_summary.txt',
			uploadedDate: '2025-06-06T09:40:00Z',
			uploadedBy: 'Jackson Martinez',
			link: '/',
			size: '20 KB'
		},
		{
			fileName: 'Client_Satisfaction_Survey.docx',
			uploadedDate: '2025-06-05T11:15:00Z',
			uploadedBy: 'Olivia Smith',
			link: '/',
			size: '380 KB'
		},
		{
			fileName: 'getting_started_guide.pdf',
			uploadedDate: '2025-06-04T13:30:00Z',
			uploadedBy: 'Liam Johnson',
			link: '/',
			size: '1.1 MB'
		},
		{
			fileName: 'Platform_Feature_Overview.pptx',
			uploadedDate: '2025-06-03T10:00:00Z',
			uploadedBy: 'Ava Davis',
			link: '/',
			size: '2.5 MB'
		},
		{
			fileName: 'feedback_form_template.docx',
			uploadedDate: '2025-06-02T16:45:00Z',
			uploadedBy: 'Noah Garcia',
			link: '/',
			size: '300 KB'
		},
		{
			fileName: 'Freelancer_Skill_Matrix.xlsx',
			uploadedDate: '2025-06-01T09:00:00Z',
			uploadedBy: 'Sophia Rodriguez',
			link: '/',
			size: '520 KB'
		},
		{
			fileName: 'general_inquiry_responses.txt',
			uploadedDate: '2025-05-31T14:20:00Z',
			uploadedBy: 'Jackson Martinez',
			link: '/',
			size: '18 KB'
		},
		{
			fileName: 'Onboarding_Resources_Clients.zip',
			uploadedDate: '2025-05-30T11:00:00Z',
			uploadedBy: 'Olivia Smith',
			link: '/',
			size: '4.8 MB'
		},
		{
			fileName: 'quick_reference_card.pdf',
			uploadedDate: '2025-05-29T10:15:00Z',
			uploadedBy: 'Liam Johnson',
			link: '/',
			size: '750 KB'
		},
		{
			fileName: 'Community_Forum_Rules.pdf',
			uploadedDate: '2025-05-28T16:00:00Z',
			uploadedBy: 'Ava Davis',
			link: '/',
			size: '1.3 MB'
		},
		{
			fileName: 'tips_for_effective_chat.docx',
			uploadedDate: '2025-05-27T09:00:00Z',
			uploadedBy: 'Noah Garcia',
			link: '/',
			size: '360 KB'
		},
		{
			fileName: 'Platform_Announcements.pptx',
			uploadedDate: '2025-05-26T12:30:00Z',
			uploadedBy: 'Sophia Rodriguez',
			link: '/',
			size: '2.7 MB'
		},
		{
			fileName: 'billing_information_guide.pdf',
			uploadedDate: '2025-05-25T14:50:00Z',
			uploadedBy: 'Jackson Martinez',
			link: '/',
			size: '1.0 MB'
		},
		{
			fileName: 'Support_Contact_Details.txt',
			uploadedDate: '2025-05-24T10:10:00Z',
			uploadedBy: 'Olivia Smith',
			link: '/',
			size: '12 KB'
		},
		{
			fileName: 'dispute_resolution_process.pdf',
			uploadedDate: '2025-05-23T11:00:00Z',
			uploadedBy: 'Liam Johnson',
			link: '/',
			size: '1.1 MB'
		},
		{
			fileName: 'Security_Best_Practices.pptx',
			uploadedDate: '2025-05-22T15:30:00Z',
			uploadedBy: 'Ava Davis',
			link: '/',
			size: '3.0 MB'
		},
		{
			fileName: 'invoice_history_template.xlsx',
			uploadedDate: '2025-05-21T09:20:00Z',
			uploadedBy: 'Noah Garcia',
			link: '/',
			size: '480 KB'
		},
		{
			fileName: 'System_Requirements_Checklist.pdf',
			uploadedDate: '2025-05-20T13:45:00Z',
			uploadedBy: 'Sophia Rodriguez',
			link: '/',
			size: '950 KB'
		},
		{
			fileName: 'user_agreement_updates.docx',
			uploadedDate: '2025-05-19T10:00:00Z',
			uploadedBy: 'Jackson Martinez',
			link: '/',
			size: '400 KB'
		}
	];
</script>

{#snippet kanbanCard(priority: number)}
	<div class="rounded bg-neutral-50 p-6 dark:bg-neutral-900">
		<div class="flex items-start justify-between text-xs">
			<div class="space-x-2">
				<!-- <p class="inline-flex items-center gap-1 text-red-800 dark:text-red-200">
					<span class="size-2 rounded-full bg-red-700 dark:bg-red-500"></span>
					High Priority
					</p> -->
				{#if priority == 1}
					<p class="inline-flex items-center gap-1 text-red-800 dark:text-red-200">
						<CellSignalFull weight="duotone" size={14} class="text-rose-600 dark:text-rose-400" />
						High Priority
					</p>
				{:else if priority == 2}
					<p class="inline-flex items-center gap-2 text-amber-800 dark:text-amber-200">
						<CellSignalMedium
							weight="duotone"
							size={14}
							class="text-yellow-600 dark:text-amber-500"
						/>
						Medium Priority
					</p>
				{:else}
					<p class="inline-flex items-center gap-1 text-emerald-800 dark:text-emerald-200">
						<CellSignalLow
							weight="duotone"
							size={14}
							class="text-emerald-600 dark:text-emerald-500"
						/>
						Low Priority
					</p>
				{/if}
			</div>
			<div>
				<p>Added on July 10, 2025</p>
				<p class="text-neutral-4400 mt-0.5 text-right text-xs">1 Day Estimated</p>
			</div>
		</div>
		<p class="mt-4 text-sm font-bold">Design Homepage Layout</p>
		<p class="mt-0.5 text-xs text-neutral-700 dark:text-neutral-400">
			Create a modern, responsive homepage design for the client’s website. Include hero banner,
			navigation, and footer sections. Follow the brand guidelines provided.
		</p>
		<div class="mt-6">
			<div>
				<p class="text-neutral-00 text-xs">Assigned To</p>

				<p class="text-sm">Ethan Caldwell, Olivia Hart</p>
			</div>
		</div>
	</div>
{/snippet}
<FullScreenDialog bind:open {close}>
	{#if isLoading}
		<Spinner />
	{:else if channel}
		<div class="grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto px-5">
			<div class="flex-none">
				<div class="container mx-auto">
					<p class="font-bold">Kanban Board of {channel.name}</p>
				</div>
			</div>

			<div
				class="container mx-auto grid h-full auto-rows-[min-content_1fr] grid-cols-3 gap-2 overflow-y-auto"
			>
				<p class="border-b py-3.5 text-center text-sm">Backlog</p>
				<p class="border-b py-3.5 text-center text-sm">In-Progress</p>
				<p class="border-b py-3.5 text-center text-sm">Completed</p>

				<div class="h-full space-y-2 overflow-y-auto">
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
				</div>
				<div class="h-full space-y-2 overflow-y-auto">
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
				</div>
				<div class="h-full space-y-2 overflow-y-auto">
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
					{@render kanbanCard(Math.floor(Math.random() * 3) + 1)}
				</div>
			</div>

			<!-- <div class="container mx-auto grid h-full grid-cols-3 gap-4 overflow-y-auto">
				<div class="max-h-full space-y-2 overflow-y-auto">
					<p class="sticky top-0 border-b bg-white py-2 text-center text-sm dark:bg-neutral-950">
						Backlog
					</p>
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
				</div>
				<div class="max-h-full space-y-2 overflow-y-auto">
					<p class="sticky top-0 border-b bg-white py-2 text-center text-sm dark:bg-neutral-950">
						In-Progress
					</p>
					{@render kanbanCard()}
				</div>
				<div class="max-h-full space-y-2 overflow-y-auto">
					<p class="sticky top-0 border-b bg-white py-2 text-center text-sm dark:bg-neutral-950">
						Completed
					</p>
					{@render kanbanCard()}
					{@render kanbanCard()}
					{@render kanbanCard()}
				</div>
			</div> -->
		</div>
	{:else}
		<ErrorMessage variant="info" text="Channel Not Selected" />
	{/if}
</FullScreenDialog>
