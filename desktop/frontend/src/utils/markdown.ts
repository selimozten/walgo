/**
 * Simple markdown to HTML converter for preview
 * This is a basic implementation for preview purposes
 * For production, consider using a library like marked or remark
 */
export const renderMarkdown = (content: string): string => {
  return content
    .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mt-4 mb-2">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 class="text-xl font-semibold mt-6 mb-3">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mt-8 mb-4">$1</h1>')
    .replace(/\*\*(.*?)\*\*/gim, '<strong class="font-semibold">$1</strong>')
    .replace(/\*(.*?)\*/gim, '<em class="italic">$1</em>')
    .replace(/!\[(.*?)\]\((.*?)\)/gim, '<img alt="$1" src="$2" class="max-w-full h-auto my-4" />')
    .replace(/\[(.*?)\]\((.*?)\)/gim, '<a href="$2" class="text-accent hover:underline">$1</a>')
    .replace(/\n/gim, '<br>');
};
