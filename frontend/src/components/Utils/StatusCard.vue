<!-- StatusCard.vue -->
<template>
  <div class="space-y-2">
    <h3 class="text-sm font-medium text-gray-400">{{ title }}</h3>
    <p class="text-base text-gray-200 break-all">{{ value || "-" }}</p>
  </div>
</template>

<script>
export default {
  name: "StatusCard",
  props: {
    title: {
      type: String,
      required: true,
    },
    value: {
      type: String,
      default: "",
    },
  },
};
</script>
