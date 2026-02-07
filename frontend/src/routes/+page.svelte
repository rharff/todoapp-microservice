<script lang="ts">
	import { onMount } from 'svelte';
	import { Activity, ArrowLeft, ArrowRight, ClipboardList, Plus, RefreshCw } from '@lucide/svelte';
	import { env } from '$env/dynamic/public';
	import type { AuditLog, Task } from '$lib/types';

	const TASK_SERVICE_URL = (env.PUBLIC_TASK_SERVICE_URL || 'http://localhost:8080').replace(/\/$/, '');
	const AUDIT_SERVICE_URL = (env.PUBLIC_AUDIT_SERVICE_URL || 'http://localhost:8081').replace(/\/$/, '');

	const stages = [
		{ key: 'todo', label: 'To Do' },
		{ key: 'in_progress', label: 'In Progress' },
		{ key: 'review', label: 'Review' },
		{ key: 'done', label: 'Done' }
	] as const;

	let activeTab: 'board' | 'audit' = 'board';
	let tasks: Task[] = [];
	let logs: AuditLog[] = [];
	let newTitle = '';
	let newStage: Task['stage'] = 'todo';
	let loadingTasks = false;
	let loadingLogs = false;
	let errorTasks: string | null = null;
	let errorLogs: string | null = null;
	let hasLoadedLogs = false;

	const formatDate = (value: string) => new Date(value).toLocaleString();

	const fetchTasks = async () => {
		loadingTasks = true;
		errorTasks = null;
		try {
			const response = await fetch(`${TASK_SERVICE_URL}/tasks`);
			if (!response.ok) throw new Error(`Task service error: ${response.status}`);
			tasks = await response.json();
		} catch (error) {
			errorTasks = error instanceof Error ? error.message : 'Unable to load tasks.';
		} finally {
			loadingTasks = false;
		}
	};

	const fetchLogs = async () => {
		loadingLogs = true;
		errorLogs = null;
		try {
			const response = await fetch(`${AUDIT_SERVICE_URL}/logs`);
			if (!response.ok) throw new Error(`Audit service error: ${response.status}`);
			logs = await response.json();
			hasLoadedLogs = true;
		} catch (error) {
			errorLogs = error instanceof Error ? error.message : 'Unable to load logs.';
		} finally {
			loadingLogs = false;
		}
	};

	const createTask = async () => {
		if (!newTitle.trim()) return;
		try {
			const response = await fetch(`${TASK_SERVICE_URL}/tasks`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ title: newTitle.trim(), stage: newStage })
			});
			if (!response.ok) throw new Error('Failed to create task.');
			newTitle = '';
			newStage = 'todo';
			await fetchTasks();
		} catch (error) {
			errorTasks = error instanceof Error ? error.message : 'Unable to create task.';
		}
	};

	const updateTask = async (taskId: string, payload: Partial<Pick<Task, 'stage' | 'position' | 'title'>>) => {
		try {
			const response = await fetch(`${TASK_SERVICE_URL}/tasks/${taskId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (!response.ok) throw new Error('Failed to update task.');
			await fetchTasks();
		} catch (error) {
			errorTasks = error instanceof Error ? error.message : 'Unable to update task.';
		}
	};

	const moveTask = (task: Task, direction: 'left' | 'right') => {
		const index = stages.findIndex((stage) => stage.key === task.stage);
		const nextIndex = direction === 'left' ? index - 1 : index + 1;
		if (nextIndex < 0 || nextIndex >= stages.length) return;
		const nextStage = stages[nextIndex].key;
		updateTask(task.id, { stage: nextStage });
	};

	const tasksByStage = (stageKey: Task['stage']) =>
		tasks
			.filter((task) => task.stage === stageKey)
			.sort((a, b) => a.position - b.position || new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

	onMount(() => {
		fetchTasks();
	});

	$: if (activeTab === 'audit' && !hasLoadedLogs && !loadingLogs) {
		fetchLogs();
	}
</script>

<div class="min-h-screen bg-background text-foreground">
	<div class="mx-auto flex w-full max-w-7xl flex-col gap-8 px-6 py-10">
		<header class="flex flex-col gap-4 rounded-3xl border border-border/70 bg-card/60 p-6 shadow-sm backdrop-blur">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div class="flex items-center gap-3">
					<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
						<ClipboardList class="h-6 w-6" />
					</div>
					<div>
						<p class="text-sm font-medium text-muted-foreground">TaskFlow Kanban</p>
						<h1 class="text-2xl font-semibold">Product-ready Task Board</h1>
					</div>
				</div>
				<button
					class="inline-flex items-center gap-2 rounded-full border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted"
					on:click={fetchTasks}
					disabled={loadingTasks}
				>
					<RefreshCw class={`h-3.5 w-3.5 ${loadingTasks ? 'animate-spin' : ''}`} />
					Refresh
				</button>
			</div>
			<p class="max-w-3xl text-sm text-muted-foreground">
				Drag-style flow without the drag: use the arrows to move tasks across stages while the services keep audit history.
			</p>
			<div class="flex flex-wrap gap-2">
				<button
					class={`rounded-full px-4 py-2 text-sm font-medium transition ${
						activeTab === 'board' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/70'
					}`}
					on:click={() => (activeTab = 'board')}
				>
					Board
				</button>
				<button
					class={`rounded-full px-4 py-2 text-sm font-medium transition ${
						activeTab === 'audit' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/70'
					}`}
					on:click={() => (activeTab = 'audit')}
				>
					Audit Logs
				</button>
			</div>
		</header>

		{#if activeTab === 'board'}
			<section class="grid gap-6 lg:grid-cols-[3fr_1fr]">
				<div class="rounded-3xl border border-border/70 bg-card p-6 shadow-sm">
					{#if errorTasks}
						<div class="mb-4 rounded-2xl border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
							{errorTasks}
						</div>
					{/if}

					<div class="grid gap-4 lg:grid-cols-4">
						{#each stages as stage}
							<div class="flex flex-col gap-3 rounded-2xl border border-border/70 bg-background/50 p-4">
								<div class="flex items-center justify-between">
									<div>
										<p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{stage.label}</p>
										<p class="text-sm text-muted-foreground">{tasksByStage(stage.key).length} tasks</p>
									</div>
								</div>
								<div class="flex flex-col gap-3">
									{#if loadingTasks}
										<p class="text-xs text-muted-foreground">Loading...</p>
									{:else if tasksByStage(stage.key).length === 0}
										<div class="rounded-xl border border-dashed border-border/80 p-4 text-center text-xs text-muted-foreground">
											No tasks
										</div>
									{:else}
										{#each tasksByStage(stage.key) as task}
											<div class="rounded-2xl border border-border/70 bg-card p-3 shadow-sm">
												<p class="text-sm font-semibold">{task.title}</p>
												<p class="mt-1 text-xs text-muted-foreground">Created {formatDate(task.created_at)}</p>
												<div class="mt-3 flex items-center justify-between">
													<div class="text-[10px] font-medium uppercase text-muted-foreground">{stage.label}</div>
													<div class="flex items-center gap-1">
														<button
															class="rounded-full border border-border p-1 text-muted-foreground hover:bg-muted"
															on:click={() => moveTask(task, 'left')}
															disabled={stage.key === 'todo'}
														>
															<ArrowLeft class="h-3 w-3" />
														</button>
														<button
															class="rounded-full border border-border p-1 text-muted-foreground hover:bg-muted"
															on:click={() => moveTask(task, 'right')}
															disabled={stage.key === 'done'}
														>
															<ArrowRight class="h-3 w-3" />
														</button>
													</div>
												</div>
											</div>
										{/each}
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>

				<div class="rounded-3xl border border-border/70 bg-card p-6 shadow-sm">
					<div class="flex items-center gap-2">
						<Plus class="h-4 w-4" />
						<h3 class="text-lg font-semibold">Create Task</h3>
					</div>
					<p class="mt-2 text-sm text-muted-foreground">Add a new task to the board.</p>
					<div class="mt-4 flex flex-col gap-3">
						<input
							class="h-11 rounded-2xl border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
							placeholder="Write a short task title"
							bind:value={newTitle}
							on:keydown={(event) => event.key === 'Enter' && createTask()}
						/>
						<select
							class="h-11 rounded-2xl border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
							bind:value={newStage}
						>
							{#each stages as stage}
								<option value={stage.key}>{stage.label}</option>
							{/each}
						</select>
						<button
							class="inline-flex items-center justify-center gap-2 rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
							on:click={createTask}
						>
							<Plus class="h-4 w-4" />
							Add task
						</button>
						<p class="text-xs text-muted-foreground">
							Tasks are persisted in the Task Service database and mirrored to the Audit Service.
						</p>
					</div>
				</div>
			</section>
		{:else}
			<section class="rounded-3xl border border-border/70 bg-card p-6 shadow-sm">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div class="flex items-center gap-2">
						<Activity class="h-4 w-4" />
						<h2 class="text-lg font-semibold">Audit Stream</h2>
					</div>
					<button
						class="inline-flex items-center gap-2 rounded-full border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted"
						on:click={fetchLogs}
						disabled={loadingLogs}
					>
						<RefreshCw class={`h-3.5 w-3.5 ${loadingLogs ? 'animate-spin' : ''}`} />
						Refresh
					</button>
				</div>

				{#if errorLogs}
					<div class="mt-4 rounded-2xl border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
						{errorLogs}
					</div>
				{/if}

				<div class="mt-6 overflow-hidden rounded-2xl border border-border/70">
					<table class="w-full text-left text-sm">
						<thead class="bg-muted/60 text-xs uppercase tracking-wide text-muted-foreground">
							<tr>
								<th class="px-4 py-3">Event</th>
								<th class="px-4 py-3">Task</th>
								<th class="px-4 py-3">Timestamp</th>
							</tr>
						</thead>
						<tbody>
							{#if loadingLogs}
								<tr>
									<td class="px-4 py-4 text-muted-foreground" colspan="3">Loading audit logs...</td>
								</tr>
							{:else if logs.length === 0}
								<tr>
									<td class="px-4 py-4 text-muted-foreground" colspan="3">No audit activity yet.</td>
								</tr>
							{:else}
								{#each logs as log}
									<tr class="border-t border-border/60">
										<td class="px-4 py-3 font-medium">{log.action_string}</td>
										<td class="px-4 py-3 text-muted-foreground">{log.task_id ?? '—'}</td>
										<td class="px-4 py-3 text-muted-foreground">{formatDate(log.created_at)}</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
			</section>
		{/if}
	</div>
</div>
