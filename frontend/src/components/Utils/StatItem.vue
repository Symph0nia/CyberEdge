<template>
  <div :class="[' ', bgColor]">
    <h3 >{{ title }}</h3>
    <p >{{ value }}</p>
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
