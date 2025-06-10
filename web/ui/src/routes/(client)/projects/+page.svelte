<script>
	import ProjectCard from '$lib/components/ProjectCard.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
</script>

<svelte:head>
	<title>Projects</title>
</svelte:head>

<div class="mx-auto lg:container">
	<div class="max-w-96">
		<SearchBar />
	</div>
</div>
<div class="mx-auto mt-6 grid grid-cols-4 gap-4 lg:container">
	<ProjectCard
		data={{
			id: 1,
			lastUpdated: new Date().toLocaleString(),
			projectName: 'Atlas CRM Integration',
			status: 'started',
			tasks: { completed: 25, inProgress: 1, total: 50 }
		}}
	/>
</div>
