<template>
  <div>
    <el-row :gutter="15" style="margin-top: 15px">
      <el-col :span="20" :offset="2">
        <el-input
          placeholder="Поиск по ФИО или СНИЛС"
          prefix-icon="el-icon-search"
          style="width: 100%"
          class="inline-input"
          v-model="searchStr"
        >
        </el-input>
      </el-col>
    </el-row>
    <el-row :gutter="15" style="margin-top: 15px">
      <el-col :span="20" :offset="2">
        <el-table
          v-loading.fullscreen.lock="loading"
          element-loading-text="Loading..."
          element-loading-spinner="el-icon-loading"
          element-loading-background="rgba(0, 0, 0, 0.8)"
          :data="retirees"
          border
          style="width: 100%"
        >
          <el-table-column type="expand">
            <template #default="props">
              <el-table
                :data="props.row['sale_coupons']"
                stripe
                style="width: 100%"
                size="mini"
              >
                <el-table-column label="Date" width="180">
                  <template #default="props">
                    {{ moment(props.row.date).format("LL") }}
                  </template>
                </el-table-column>
                <el-table-column prop="count" label="Количество" width="180">
                </el-table-column>
                <el-table-column prop="color" label="Цвет" width="180"> </el-table-column>
                <el-table-column prop="cashier" label="Кассир"> </el-table-column>
              </el-table>
            </template>
          </el-table-column>
          <el-table-column prop="full_name" label="Ф.И.О."> </el-table-column>
          <el-table-column prop="snils" label="СНИЛС" width="200"> </el-table-column>
          <el-table-column label="Дата рождения" width="200">
            <template #default="props">
              {{ moment(props.row['birthdate']).format("LL") }}
            </template>
          </el-table-column>
        </el-table>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import RetireesDataService from "../services/RetireesDataService";
import moment from "moment";

export default {
  name: "retirees-list",
  data() {
    return {
      loading: true,
      retirees: [],
      searchStr: "",
    };
  },
  created: function () {
    this.moment = moment;
  },
  watch: {
    searchStr(newStr) {
      if (newStr.length > 2) {
        this.searchRetirees(newStr);
      } else {
        this.retrieveRetirees();
      }
    },
  },
  methods: {
    retrieveRetirees() {
      RetireesDataService.getAll()
        .then((response) => {
          this.retirees = response.data;
          console.log(response.data);
          this.loading = false;
        })
        .catch((e) => {
          console.log(e);
        });
    },
    searchRetirees(searchStr) {
      RetireesDataService.find(searchStr)
        .then((response) => {
          this.retirees = response.data;
          console.log(response.data);
        })
        .catch((e) => {
          console.log(e);
        });
    },
  },
  mounted() {
    this.retrieveRetirees();
  },
};
</script>
