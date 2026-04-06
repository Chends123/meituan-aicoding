<template>
  <el-tabs :value="value" @input="$emit('input', $event)">
    <el-tab-pane label="全部" name="all" />
    <el-tab-pane label="好评" name="positive" />
    <el-tab-pane label="差评" name="negative" />
  </el-tabs>
</template>

<script>
export default {
  name: "ReviewFilterTabs",
  props: {
    value: {
      type: String,
      default: "all"
    }
  }
};
</script>

<style scoped>
::v-deep .el-tabs__active-bar {
  background-color: var(--mt-primary);
}

::v-deep .el-tabs__item.is-active {
  color: var(--mt-primary-deep);
}

::v-deep .el-tabs__item:hover {
  color: var(--mt-primary);
}

::v-deep .el-tabs__nav-wrap::after {
  background-color: var(--mt-border);
}
</style>
