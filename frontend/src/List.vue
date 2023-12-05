<script setup>
import { ref, computed, watch } from "vue";
import { formatDateTime, makeGetRequest } from "./utils.js";
import Create from "./Create.vue";

let limit = 10;
let offset = ref(0);
let total = ref(0);
let busy = ref(true);

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
    const data = await makeGetRequest(`/private/api/links?limit=${limit}&offset=${offset.value}`, busy);
    links.value = data.results;
    total.value = data.total;
}

watch(offset, fetchLinks, { immediate: true });
</script>

<template>
    <article>
        <Create @closed="fetchLinks((offset = 0))" />
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
                :aria-busy="busy && activePage === i"
            >
                {{ i }}
            </a>
        </footer>
    </article>
</template>

<style scoped>
article > footer > a {
    text-decoration: none;
    padding: 0.3rem 0.7rem;
}
</style>
