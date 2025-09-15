<template>
  <div :class="['p-4 rounded-lg shadow', bgColor]">
    <h3 class="text-xl font-bold">{{ title }}</h3>
    <p class="text-3xl">{{ value }}</p>
  </div>
</template>

<script>
export default {
  name: "StatItem",
  props: {
    title: String,
    value: [String, Number],
    bgColor: String,
  },
};
</script>
