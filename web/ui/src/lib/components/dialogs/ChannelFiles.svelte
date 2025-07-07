<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import DownloadIcon from '$lib/components/icons/DownloadIcon.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
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

<FullScreenDialog bind:open {close}>
	{#if isLoading}
		<Spinner />
	{:else if channel}
		<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
			<div class="flex-none">
				<div class="container mx-auto flex items-end justify-between gap-4">
					<p class="font-bold">Uploaded Files in {channel.name}</p>
					<div class="w-full max-w-xs">
						<SearchBar />
					</div>
				</div>
			</div>

			<div class="overflow-y-auto">
				{#if files.length > 0}
					<Table.Root class="container mx-auto">
						<Table.Header>
							<Table.Row>
								<Table.Head class="font-bold">File Name</Table.Head>
								<Table.Head class="font-bold">Size</Table.Head>
								<Table.Head class="font-bold">Uploaded Date</Table.Head>
								<Table.Head class="font-bold">Uploaded By</Table.Head>
								<Table.Head class="font-bold">Actions</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each files as file}
								<Table.Row>
									<Table.Cell>{file.fileName}</Table.Cell>
									<Table.Cell>{file.size}</Table.Cell>
									<Table.Cell>{file.uploadedDate}</Table.Cell>
									<Table.Cell>{file.uploadedBy}</Table.Cell>
									<Table.Cell>
										<a
											href={file.link}
											class="block w-fit cursor-pointer px-2 text-neutral-600
										transition hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-neutral-50"
										>
											<DownloadIcon class="size-4.5 " />
										</a>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				{/if}
			</div>
			{#if files.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page />
				</div>
			{/if}
		</div>
	{:else}
		<ErrorMessage variant="info" text="Channel Not Selected" />
	{/if}
</FullScreenDialog>
