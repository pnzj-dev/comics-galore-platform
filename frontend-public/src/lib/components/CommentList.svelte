<script lang="ts">
	import CommentList from '$lib/components/CommentList.svelte';
	import CommentForm from '$lib/components/CommentForm.svelte';

	export interface Comment {
		id: string;
		comic_id: string;
		user_id: string;
		parent_id?: string;
		body_text: string;
		created_at: string;
		replies?: Comment[];
	}

	interface Props {
		comments: Comment[];
		onReply: (commentId: string) => void;
		onDelete: (commentId: string) => void;
		onSubmitComment: (bodyText: string, parentId?: string) => Promise<void>;
		userId?: string;
		role?: string;
		depth?: number;
	}

	let { comments, onReply, onDelete, onSubmitComment, userId, role = 'user', depth = 0 }: Props = $props();

	let activeReply = $state<string | null>(null);

	function formatDate(d: string): string {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
	}

	function canDelete(commentUserId: string): boolean {
		return userId === commentUserId || role === 'admin' || role === 'moderator';
	}

	async function handleReply(bodyText: string) {
		await onSubmitComment(bodyText, activeReply!);
		activeReply = null;
	}
</script>

{#if comments.length > 0}
	<div class="space-y-3 {depth > 0 ? 'ml-6 border-l-2 border-border pl-4' : ''}">
		{#each comments as comment (comment.id)}
			<div class="group">
				<div class="flex items-start gap-2">
					<div class="size-7 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-[10px] font-medium text-purple-600 dark:text-purple-300 shrink-0 mt-0.5">
						{comment.user_id.charAt(0).toUpperCase()}
					</div>
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2 flex-wrap">
							<span class="text-xs font-medium">User</span>
							<span class="text-[10px] text-muted-foreground">{formatDate(comment.created_at)}</span>
							{#if canDelete(comment.user_id)}
								<button onclick={() => onDelete(comment.id)} class="text-[10px] text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity" aria-label="Delete comment">Delete</button>
							{/if}
						</div>
						<p class="text-sm mt-0.5 leading-relaxed">{comment.body_text}</p>
						<button
							onclick={() => { onReply(comment.id); activeReply = activeReply === comment.id ? null : comment.id; }}
							class="text-[10px] text-muted-foreground hover:text-primary mt-1"
						>
							{activeReply === comment.id ? 'Cancel' : 'Reply'}
						</button>

						{#if activeReply === comment.id}
							<div class="mt-2">
								<CommentForm onSubmit={handleReply} placeholder="Write a reply..." />
							</div>
						{/if}
					</div>
				</div>

				{#if comment.replies && comment.replies.length > 0}
					<CommentList comments={comment.replies} {onReply} {onDelete} {onSubmitComment} {userId} {role} depth={depth + 1} />
				{/if}
			</div>
		{/each}
	</div>
{/if}
