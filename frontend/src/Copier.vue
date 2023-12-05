<script setup>
import { onBeforeMount, ref } from "vue";

const props = defineProps({ text: String });
let status = ref("pending");

async function clipboardHandler() {
    try {
        await navigator.clipboard.writeText(props.text);
        status.value = "copied";
    } catch (_) {
        status.value = "failed";
    }
}

onBeforeMount(() => {
    if (!("clipboard" in navigator)) status.value = "disabled";
});
</script>

<template>
    <kbd v-if="status === 'pending'" @click.prevent="clipboardHandler">Copy</kbd>
    <ins v-if="status === 'copied'">Copied</ins>
    <span v-if="status === 'failed'">Failed</span>
</template>

<style scoped>
kbd {
    cursor: pointer;
    margin: -0.2rem 0 -0.2rem;
}
kbd:hover {
    font-weight: normal;
}
</style>
