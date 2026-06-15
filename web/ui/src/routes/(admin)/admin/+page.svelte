<script lang="ts">
	import ProjectsCreatedMetricsCard from '$lib/components/ProjectsCreatedMetricsCard.svelte';
	import RecentProjectsCard from '$lib/components/RecentProjectsCard.svelte';
	import RecentContractsCard from '$lib/components/RecentContractsCard.svelte';
	import RecentInvoicesCard from '$lib/components/RecentInvoicesCard.svelte';
	import InvoicesOverviewCard from '$lib/components/InvoicesOverviewCard.svelte';
	import InvoicesPaidMetricsCard from '$lib/components/InvoicesPaidMetricsCard.svelte';
</script>

<svelte:head>
	<title>Dashboard - Administrator</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<ProjectsCreatedMetricsCard class="col-span-2 h-80" />
	<RecentProjectsCard role="administrator" class="col-span-2 h-80" />
	<RecentContractsCard role="administrator" class="col-span-4 max-h-100 min-h-80" />

	<div class="col-span-4 grid grid-cols-5 gap-2">
		<RecentInvoicesCard role="administrator" class="col-span-2 max-h-80 min-h-50" />
		<InvoicesOverviewCard class="max-h-80 min-h-50" />
		<InvoicesPaidMetricsCard class="col-span-2 max-h-80 min-h-50" />
	</div>
</div>
