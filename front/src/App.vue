<script setup>
import { RouterLink, RouterView } from "vue-router";
</script>

<template>
  <header>
    <div>
      <h2>Govnocloud</h2>
      <nav>
        <RouterLink to="/nodes">Nodes</RouterLink>
        <RouterLink to="/vms">VMs</RouterLink>
        <RouterLink to="/containers">Containers</RouterLink>
        <RouterLink to="/files">Files</RouterLink>
        <RouterLink to="/k8s">K8s</RouterLink>
        <RouterLink to="/llms">LLMs</RouterLink>
        <RouterLink to="/dbs">DBs</RouterLink>
      </nav>
    </div>
  </header>
  <div id="router_body">
    <RouterView />
  </div>
</template>

<style scoped>
nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 1px solid var(--color-border);
  
}
.nav-item.selected {
  background-color: green
}

nav a:first-of-type {
  border: 0;
}

</style>