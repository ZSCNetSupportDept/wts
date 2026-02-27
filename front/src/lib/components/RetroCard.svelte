<script>
	let { children, ...rest } = $props();
</script>

<div class="card" {...rest}>
	<div>
		{@render children?.()}
	</div>
</div>

<style>
	.card {
		position: relative;
		padding: 20px;
		margin-bottom: 30px; /* Increased margin to accommodate shadow */
		margin-left: 15px;
		margin-right: 20px; /* Increased to balance the shadow offset */
		z-index: 0;
	}

	.card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: #efefef;
		border: 2px solid #000;
		z-index: -1;
		box-sizing: border-box;
	}

	.card::after {
		content: '';
		position: absolute;
		top: 10px;
		left: 10px;
		width: 100%;
		height: 100%;
		background-color: lightblue;
		background-image: repeating-linear-gradient(
			45deg,
			#000000,
			#000000 1px,
			transparent 1px,
			transparent 5px
		);
		z-index: -2;
	}

	.card :global(p) {
		line-height: 1.6;
		color: #4a4949;
	}

	.card :global(h2) {
		font-size: 25px;
	}
</style>
