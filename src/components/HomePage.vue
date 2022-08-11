<template>
  <div>
    <el-autocomplete
      class="inline-input"
      v-model="state2"
      :fetch-suggestions="querySearch"
      placeholder="Please Input"
      :trigger-on-focus="false"
      @select="handleSelect"
    ></el-autocomplete>
  </div>
</template>

<script>
import RetireesDataService from "../services/RetireesDataService";

export default {
  name: "home-page",
  data() {
    return {
      retirees: [],
      retiree: "",
    };
  },
  methods: {
    loadAll() {},
    querySearch(queryString, cb) {
      var results = (queryString.length > 2)
        ? this.retirees.filter(this.createFilter(queryString))
        : [];
      // call callback function to return suggestions
      cb(results);
    },
    createFilter() {
        RetireesDataService.getAll()
        .then((response) => {
          this.retirees = response.data;
          console.log(response.data);
          this.loading = false;
        })
        .catch((e) => {
          console.log(e);
        });
    }
  },
};
</script>
