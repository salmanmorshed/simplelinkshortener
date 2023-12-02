<script setup>
import { ref, watch } from "vue";

let totalPages = ref(0);
let activePage = ref(1);
let links = ref([]);

let showCreateDialog = ref(false);
let createDialogURL = ref("");
let newShortLinkURL = ref("");

async function fetchLinks() {
    const response = await fetch(`/private/api/links?page=${activePage.value}`, { credentials: "include" });
    const data = await response.json();
    totalPages.value = data["total_pages"];
    links.value = data["results"];
}

async function createLink() {
    const response = await fetch("/private/api/links", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ url: createDialogURL.value }),
    });
    const data = await response.json();
    newShortLinkURL.value = data.short_url;
    createDialogURL.value = "";
}

const formatDateTime = input => new Date(Date.parse(input)).toLocaleString();

watch(activePage, async () => await fetchLinks(), { immediate: true });
watch(showCreateDialog, val => {
    if (!val) {
        createDialogURL.value = "";
        newShortLinkURL.value = "";
    }
});
</script>

<template>
    <main class="container">
        <article>
            <table role="grid">
                <thead>
                    <th>Short Link</th>
                    <th>Link</th>
                    <th>Visits</th>
                    <th>Created at</th>
                </thead>
                <tbody>
                    <tr v-for="link in links" :key="link.short_url">
                        <td>{{ link.short_url }}</td>
                        <td>{{ link.url }}</td>
                        <td>{{ link.visits }}</td>
                        <td>{{ formatDateTime(link.created_at) }}</td>
                    </tr>
                </tbody>
            </table>
            <footer>
                <a v-for="i in totalPages" :class="{ contrast: i === activePage }" @click.prevent="activePage = i">
                    {{ i }}
                </a>
            </footer>
        </article>
        <button type="button" class="open" @click="showCreateDialog = true">New Short Link</button>
    </main>

    <dialog :open="showCreateDialog">
        <article style="width: 50%">
            <header>
                <a class="close" @click.prevent="showCreateDialog = false"></a>
                <div>New Short Link</div>
            </header>
            <form @submit.prevent="createLink()">
                <input
                    type="url"
                    v-model="createDialogURL"
                    placeholder="Enter a link (e.g. https://example.com/...)"
                    required
                />
                <button type="submit">Shorten Link</button>
            </form>
            <footer :hidden="!newShortLinkURL">
                <a :href="newShortLinkURL" target="_blank">{{ newShortLinkURL }}</a>
                <ins hidden>Copied</ins>
            </footer>
        </article>
    </dialog>
</template>

<style scoped>
dialog > article > footer {
    text-align: left;
}
main > article > footer > a {
    text-decoration: none;
    padding: 0.25rem 0.75rem;
}
</style>
