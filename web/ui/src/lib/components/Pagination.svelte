<script lang="ts">
	import { Pagination } from 'bits-ui';
	import CaretLeftIcon from './icons/CaretLeftIcon.svelte';
	import CaretRightIcon from './icons/CaretRightIcon.svelte';

	let { page = $bindable() } = $props();
</script>

<Pagination.Root count={100} perPage={10} bind:page>
	{#snippet children({ pages, range })}
		<div class="flex items-center gap-4">
			<p class="text-muted-foreground text-center text-xs">
				Showing {range.start} - {range.end}
			</p>
			<div class="flex items-center">
				<Pagination.PrevButton
					class="mr-[25px] inline-flex size-7 
                    cursor-pointer items-center justify-center rounded bg-transparent 
                    hover:bg-neutral-100 active:scale-[0.98] disabled:cursor-not-allowed disabled:text-neutral-500 
                    hover:disabled:bg-transparent dark:hover:bg-neutral-900"
				>
					<CaretLeftIcon class="size-4.5" />
				</Pagination.PrevButton>

				<div class="flex items-center gap-2.5">
					{#each pages as page (page.key)}
						{#if page.type === 'ellipsis'}
							<div class="select-none text-sm font-medium text-neutral-950 dark:text-neutral-50">
								...
							</div>
						{:else}
							<Pagination.Page
								{page}
								class="data-selected:bg-neutral-950 dark:data-selected:bg-neutral-50 
                                data-selected:text-neutral-50 dark:data-selected:text-neutral-950 
                                data-selected:hover:bg-neutral-950 inline-flex size-7 cursor-pointer 
                                select-none items-center justify-center rounded bg-transparent text-sm font-medium 
                                hover:bg-neutral-100 active:scale-[0.98] disabled:cursor-not-allowed 
                                disabled:opacity-50 hover:disabled:bg-transparent dark:hover:bg-neutral-900"
							>
								{page.value}
							</Pagination.Page>
						{/if}
					{/each}
				</div>
				<Pagination.NextButton
					class="hover:bg-dark-10 disabled:text-muted-foreground ml-[29px] inline-flex size-7 items-center justify-center rounded bg-transparent active:scale-[0.98] disabled:cursor-not-allowed hover:disabled:bg-transparent"
				>
					<CaretRightIcon class="size-4.5" />
				</Pagination.NextButton>
			</div>
		</div>
	{/snippet}
</Pagination.Root>
