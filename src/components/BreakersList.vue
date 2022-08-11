<template>
  <div>
    <el-row :gutter="20" style="margin-top: 15px">
      <el-col :span="13" :offset="1">
        <el-skeleton style="width: 100%" :loading="loading" animated :rows="10">
          <el-table
            :data="tableData"
            border
            style="width: 100%"
            :row-style="tableRowClassName"
          >
            <el-table-column type="expand">
              <template #default="props">
                <el-timeline>
                  <el-timeline-item
                    v-for="(activity, index) in props.row.timeline"
                    :key="index"
                    :timestamp="moment(activity.timestamp).format('LL')"
                  >
                    {{ activity.content }}
                  </el-timeline-item>
                </el-timeline>
              </template>
            </el-table-column>
            <el-table-column label="Дата" width="120">
              <template #default="scope">
                {{ moment(scope.row.date).format("LL") }}
              </template>
            </el-table-column>
            <el-table-column prop="snils" label="СНИЛС" width="150"> </el-table-column>
            <el-table-column prop="name" label="Фамилия Имя Отчество"> </el-table-column>
            <el-table-column prop="pan" label="PAN"> </el-table-column>
            <el-table-column width="100" label="Обработан" align="center">
              <template #default="scope">
                <!-- <el-checkbox @change="checkSnils(scope.row.snils)" label="Обработан" border size="medium"></el-checkbox> -->
                <el-button
                  @click="checkSnils(scope)"
                  :type="getCheckButtonType(scope.row.checked)"
                  icon="el-icon-check"
                  size="mini"
                  circle
                ></el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-skeleton>
      </el-col>
      <el-col :span="9">
        <el-row :gutter="15" justify="space-around" align="middle">
          <el-col :span="22" :offset="1">
            <el-affix :offset="20">
              <el-card style="width: 100%; height: 100%; margin-top: 10px">
                <el-space direction="vertical" style="width: 100%">
                  <el-date-picker
                    v-model="fromDates"
                    size="medium"
                    unlink-panels
                    type="daterange"
                    range-separator="-"
                    start-placeholder="От"
                    end-placeholder="До"
                    :shortcuts="shortcuts"
                    format="DD/MM/YYYY"
                  >
                  </el-date-picker>
                  <el-checkbox
                    v-model="onlyCheckedBrackers"
                    label="Показывать только не обработанных"
                    border
                  ></el-checkbox>
                  <el-descriptions :column="1" border style="width: 100%">
                    <el-descriptions-item>
                      <template #label>
                        <i class="el-icon-user"></i>
                        Нарушителей
                      </template>
                      {{ tableData.length }}
                    </el-descriptions-item>
                  </el-descriptions>
                  <el-button
                    type="primary"
                    icon="el-icon-download"
                    plain
                    @click="saveToExcel"
                    >Сохранить в Excel</el-button
                  >
                </el-space>
              </el-card>
            </el-affix>
          </el-col>
        </el-row>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import BreakersDataService from "../services/BreakersDataService";
import moment from "moment";
export default {
  name: "breakers-list",
  data() {
    return {
      loading: true,
      breakers: [],
      onlyCheckedBrackers: false,
      fromDates: null,
      shortcuts: [
        {
          text: "Сегодня",
          value: (() => {
            const end = new Date();
            const start = new Date();
            return [start, end];
          })(),
        },
        {
          text: "2 дня",
          value: (() => {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 1);
            return [start, end];
          })(),
        },
        {
          text: "Неделя",
          value: (() => {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
            return [start, end];
          })(),
        },
      ],
    };
  },
  methods: {
    saveToExcel() {
      console.log(this.tableData);
      BreakersDataService.saveToExcel(this.tableData)
        .then((response) => {
          let blob = new Blob([response.data], {
            type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          });
          let url = window.URL.createObjectURL(blob);
          let a = document.createElement("a");
          a.href = url;
          a.download = "Нарушители.xlsx";
          a.click();
          window.URL.revokeObjectURL(url);
        })
        .catch((e) => {
          console.log(e);
        });
    },
    getCheckButtonType(ch) {
      return ch === "false" ? "info" : "success";
    },
    checkSnils(scope) {
      BreakersDataService.check(scope.row.snils)
        .then((response) => {
          console.log(response.data);
          if (response.data[scope.row.snils] === "true") {
            scope.row.checked = true;
          } else if (response.data[scope.row.snils] === "false") {
            scope.row.checked = "false";
          }
          console.log(scope.row.checked);
        })
        .catch((e) => {
          console.log(e);
        });
    },
    tableRowClassName({ row }) {
      return row.checked == "false" ? "background: #fdf6ec; border-color: #f5dab1;" : "";
    },
    breakersRetirees() {
      BreakersDataService.getAll()
        .then((response) => {
          this.breakers = response.data;
          this.loading = false;
        })
        .catch((e) => {
          console.log(e);
        });
    },
  },
  mounted() {
    this.breakersRetirees();
  },
  created: function () {
    this.moment = moment;
  },
  computed: {
    tableData: function () {
      var temp = this.breakers;

      if (this.fromDates) {
        temp = temp.filter((data) =>
          this.moment(data.date).isBetween(
            this.fromDates[0],
            this.fromDates[1],
            "day",
            "[]"
          )
        );
      }

      if (this.onlyCheckedBrackers) {
        return temp.filter((data) => data.checked == "false");
      } else {
        return temp;
      }
    },
  },
};
</script>
