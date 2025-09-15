<!-- StatusCard.vue -->
<template>
  <div >
    <h3 >{{ title }}</h3>
    <p >{{ value || "-" }}</p>
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
