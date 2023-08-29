<script>
	import { enhance } from '$app/forms';
	import { Button, Textarea } from 'flowbite-svelte';
</script>

<div class="flex flex-row justify-center">
	<main class="w-144 pt-2">
		<header class="sticky">
			<h1 class="text-2xl font-bold text-primary-700">Propolis</h1>
		</header>
		<section>
			<form class="flex flex-row" method="post" action="?/post" use:enhance>
				<div class="w-24" />
				<div class="w-full flex flex-col items-end gap-2">
					<Textarea name="post" rows="5" placeholder="Fill the world with wonder&hellip;" />
					<Button type="submit">Post</Button>
				</div>
			</form>
		</section>
	</main>
</div>
