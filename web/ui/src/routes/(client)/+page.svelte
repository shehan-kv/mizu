<script>
</script>

<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<div class="mx-auto grid grid-cols-5 gap-4 lg:container">
	<div class="col-span-2 grid auto-rows-[1fr_min-content] gap-6 rounded border p-6">
		<div class="">Chart goes here</div>
		<div class="space-y-6">
			<div class="space-y-1">
				<p class="text-sm">Total Paid Amount</p>
				<p class="text-4xl">$ 143,000.00</p>
			</div>
			<div class="grid grid-cols-2">
				<div class="space-y-1">
					<p class="text-sm">Outstanding Balance</p>
					<p class="text-2xl">$ 99,000.00</p>
				</div>
				<div class="space-y-1">
					<p class="text-sm">Overdue Amount</p>
					<p class="text-2xl">$ 00.00</p>
				</div>
			</div>
		</div>
	</div>
	<div class="space-y-4"></div>
</div>

<div class="container mx-auto mt-10 border-t py-4">
	<p class="text-neutral-500">Recent Projects</p>

	<div class="mt-4 grid grid-cols-4 gap-4"></div>
</div>
