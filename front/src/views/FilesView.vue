<template>
  <div>
    <table class="bordered-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Size</th>
          <th>Node</th>
          <th>Committed</th>
          <th>Timestamp</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="file in files" :key="file.name">
          <td>{{ file.name }}</td>
          <td>{{ file.size }}</td>
          <td>{{ file.nodename }}</td>
          <td>{{ file.committed }}</td>
          <td>{{ file.timestamp }}</td>
          <td> <button @click="downloadFile(file.name)">Download</button> </td>
          <td> <button @click="removeFile(file.name)">Remove</button> </td>
        </tr>
      </tbody>
    </table>
    <input type="file" ref="fileUpload" @change="handleFileUpload" style="display: none" />
    <button @click="triggerFileUpload">Upload File</button>

  </div>
</template>

<script>
export default {
  data() {
    return {
      files: []
    };
  },
  created() {
    this.fetchFiles();
  },
  methods: {
    fetchFiles() {
      fetch('http://master.govno.cloud:6969/api/v1/files')
        .then(response => response.json())
        .then(data => {
          this.files = data;
        })
        .catch(error => {
          console.error('Error fetching files:', error);
        });
    },
    triggerFileUpload() {
      this.$refs.fileUpload.click();
    },
    handleFileUpload() {
      const file = this.$refs.fileUpload.files[0];
      
    },
  }
}
</script>

<style scoped>
.bordered-table {
  border-collapse: collapse;
  width: 100%;
}

.bordered-table th, .bordered-table td {
  border: 1px solid green;
  padding: 10px
}
</style>