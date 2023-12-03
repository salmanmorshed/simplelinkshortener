<script setup>
import { ref, computed, watch } from "vue";
import Create from "./Create.vue";

let showCreateDialog = ref(false);

let limit = 10;
let offset = ref(0);
let total = ref(0);

const totalPages = computed(() => Math.ceil(total.value / limit));
const activePage = computed({
    get() {
        return Math.floor(offset.value / limit) + 1;
    },
    set(value) {
        offset.value = (value - 1) * limit;
    },
});

let links = ref([]);

async function fetchLinks() {
    const url = `/private/api/links?limit=${limit}&offset=${offset.value}`;
    const response = await fetch(url, { credentials: "include" });
    const data = await response.json();
    links.value = data.results;
    total.value = data.total;
}

function formatDateTime(input) {
    return new Date(Date.parse(input)).toLocaleString();
}

watch(offset, fetchLinks, { immediate: true });
</script>

<template>
    <main class="container">
        <article>
            <table role="grid">
                <thead>
                    <tr>
                        <th>Slug</th>
                        <th>Link</th>
                        <th>Visits</th>
                        <th>Created at</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="link in links" :key="link.slug">
                        <td>{{ link.slug }}</td>
                        <td>{{ link.url }}</td>
                        <td>{{ link.visits }}</td>
                        <td>{{ formatDateTime(link.created_at) }}</td>
                    </tr>
                </tbody>
            </table>
            <footer>
                <a
                    href="#"
                    v-for="i in totalPages"
                    :class="{ contrast: i === activePage }"
                    @click.prevent="activePage = i"
                >
                    {{ i }}
                </a>
            </footer>
        </article>
        <button type="button" class="open" @click="showCreateDialog = true">New Short Link</button>
    </main>
    <Create v-model="showCreateDialog" />
</template>

<style scoped>
main > article > footer > a {
    text-decoration: none;
    padding: 0.3rem 0.7rem;
}
</style>
