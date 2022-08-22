<template>
  <div>
    <el-container>
      <el-header style="height: 220px">
        <el-row :gutter="15" justify="space-around" align="middle">
          <el-col :span="12">
            <el-card style="width: 100%; height: 100%; margin-top: 10px">
              <el-row :gutter="15" justify="space-between" align="middle">
                <el-col :span="20" :offset="2">
                  <el-descriptions
                    title="Реестры ЕРЦ Прогресс"
                    :column="2"
                    :size="mini"
                    border
                  >
                    <template #extra>
                      <el-button
                        icon="el-icon-refresh"
                        size="mini"
                        type="primary"
                        @click="syncERC()"
                        >Проверить</el-button
                      >
                    </template>
                    <el-descriptions-item label="Реестров">{{
                      stat.total
                    }}</el-descriptions-item>
                    <el-descriptions-item label="Покупок">{{
                      stat.sales
                    }}</el-descriptions-item>
                    <el-descriptions-item label="Продано талонов">{{
                      stat.quantity
                    }}</el-descriptions-item>
                    <el-descriptions-item label="На общую сумму">{{
                      stat.amount
                    }}</el-descriptions-item>
                    <el-descriptions-item label="Пенсионеров в базе">{{
                      stat.retirees
                    }}</el-descriptions-item>
                    <el-descriptions-item label="Ошибки в данных">
                      <el-button
                        type="warning"
                        @click="dialogTableVisible = !dialogTableVisible"
                        plain
                        icon="el-icon-document"
                        >{{errorsData ? errorsData.length : 0 }}</el-button
                      >
                    </el-descriptions-item>
                  </el-descriptions>
                </el-col>
              </el-row>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card style="width: 100%; height: 100%; margin-top: 10px">
              <el-row :gutter="15" justify="space-between" align="middle">
                <el-col :span="20" :offset="2">
                  <el-descriptions title="Реестры РСТК" :column="2" :size="mini" border>
                    <template #extra>
                      <el-upload
                        action="/api/updates/uploadRSTK"
                        accept=".txt"
                        :show-file-list="false"
                        :before-upload="uploadStart"
                        :on-success="uploadEnd"
                      >
                        <el-button icon="el-icon-document-add" size="mini" type="primary"
                          >Загрузить</el-button
                        >
                      </el-upload>
                    </template>
                    <el-descriptions-item label="Реестров">{{
                      stat.updates_rstk
                    }}</el-descriptions-item>
                    <el-descriptions-item label="Выдано карт">{{
                      stat.cards
                    }}</el-descriptions-item>
                  </el-descriptions>
                  <el-divider><i class="el-icon-star-on"></i></el-divider>
                  <el-space
                    style="width: 100%; margin-top: 15px; justify-content: space-between"
                  >
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
                    <el-button
                      type="primary"
                      icon="el-icon-download"
                      plain
                      @click="saveRSTKToExcel"
                      >Отчёт для Прогресс</el-button
                    >
                  </el-space>
                </el-col>
              </el-row>
            </el-card>
          </el-col>
        </el-row>
      </el-header>

      <!-- <el-divider><i class="el-icon-star-on"></i></el-divider> -->

      <el-main style="margin-top: 15px">
        <el-row :gutter="15">
          <el-col :span="12">
            <el-table
              v-loading.fullscreen.lock="loading"
              element-loading-text="Loading..."
              element-loading-spinner="el-icon-loading"
              element-loading-background="rgba(0, 0, 0, 0.8)"
              :data="ercUpdates"
              border
              style="width: 100%"
            >
              <el-table-column type="expand">
                <template #default="props">
                  <el-alert
                    v-for="e in props.row.incorrect"
                    :key="e"
                    :title="e.snils"
                    type="info"
                    :description="e.full_name + ' ' + moment(e.birthdate).format('LL')"
                    :closable="false"
                  >
                  </el-alert>
                </template>
              </el-table-column>
              <el-table-column label="Получен">
                <template #default="props">
                  {{ moment(props.row['datetime_received']).format("LL") }}
                </template>
              </el-table-column>
              <el-table-column label="Обработан">
                <template #default="props">
                  {{ moment(props.row['datetime_parsed']).format("LL") }}
                </template>
              </el-table-column>
              <el-table-column prop="lines" label="Покупок"> </el-table-column>
              <el-table-column prop="incorrect.length" label="Ошибок"> </el-table-column>
            </el-table>
          </el-col>

          <el-col :span="12">
            <el-table
              v-loading.fullscreen.lock="loading"
              element-loading-text="Loading..."
              element-loading-spinner="el-icon-loading"
              element-loading-background="rgba(0, 0, 0, 0.8)"
              :data="rstkUpdates"
              border
              style="width: 100%"
            >
              <el-table-column type="expand">
                <template #default="props">
                  <el-alert
                    v-for="e in props.row.errors"
                    :key="e"
                    :title="e.snils"
                    type="info"
                    :description="e.full_name"
                    :closable="false"
                  >
                  </el-alert>
                </template>
              </el-table-column>
              <el-table-column prop="datetime_received" label="За дату">
                <template #default="props">
                  {{ moment(props.row['from_date']).format("LL") }}
                </template>
              </el-table-column>
              <el-table-column label="Тип">
                <template #default="props">
                  {{ props.row['type_id'] === 2 ? "МИР" : "СТК" }}
                </template>
              </el-table-column>
              <el-table-column prop="datetime_parse" label="Обработан">
                <template #default="props">
                  {{ moment(props.row['uploaded_at']).format("LL") }}
                </template>
              </el-table-column>
              <el-table-column prop="lines" label="Выдано карт"> </el-table-column>
              <el-table-column prop="errors.length" label="Ошибок"> </el-table-column>
              <el-table-column>
                <template #default="scope">
                  <el-popconfirm
                    title="Вы действительно хотите удалить этот реестр?"
                    confirmButtonText="Да"
                    cancelButtonText="Нет, спасибо."
                    iconColor="red"
                    confirmButtonType="text"
                    cancelButtonType="primary"
                    icon="el-icon-delete"
                    @confirm="handleDelete(scope.$index, scope.row.id)"
                  >
                    <template #reference>
                      <el-button icon="el-icon-document-delete" size="mini" type="danger">
                        Удалить
                      </el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
            </el-table>
          </el-col>
        </el-row></el-main
      >
    </el-container>
    <el-dialog
      title="Загрузка реестра РСТК"
      v-model="uploadDialogVisible"
      width="400px"
      :show-close="false"
      :close-on-press-escape="false"
      :close-on-click-modal="false"
      center
    >
      <el-result v-if="OkVisible" icon="success" title="Реестр РСТК успешно загружен">
        <template #extra>
          <el-button
            type="primary"
            @click="uploadDialogVisible = false"
            v-if="OkVisible"
            size="medium"
            >OK</el-button
          >
        </template>
      </el-result>
      <el-result v-else icon="warning" title="Ожидайте!"> </el-result>
    </el-dialog>
    <el-dialog
      title="Синхронизация реестров ЕРЦ Прогресс"
      v-model="uploadDialogVisible2"
      width="400px"
      :show-close="false"
      :close-on-press-escape="false"
      :close-on-click-modal="false"
      center
    >
      <el-result
        v-if="OkVisible2"
        icon="success"
        title="Реестры ЕРЦ Прогресс успешно синхронизированы"
      >
        <template #extra>
          <el-button
            type="primary"
            @click="uploadDialogVisible2 = false"
            v-if="OkVisible2"
            size="medium"
            >OK</el-button
          >
        </template>
      </el-result>
      <el-result v-else icon="warning" title="Ожидайте!"> </el-result>
    </el-dialog>
    <el-dialog title="Ошибочные данные" v-model="dialogTableVisible">
      <el-table :data="errorsData">
        <el-table-column property="id" label="ID" width="50"></el-table-column>
        <el-table-column property="snils" label="СНИЛС" width="150"></el-table-column>
        <el-table-column property="birthdate" label="День рождения" width="120">
          <template #default="props">
            {{ moment(props.row.birthdate).format("L") }}
          </template>
        </el-table-column>
        <el-table-column property="full_name" label="ФИО"></el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script>
import UpdatesDataService from "../services/UpdatesDataService";
import moment from "moment";

export default {
  name: "updates-list",
  data() {
    return {
      dialogTableVisible: false,
      errorsData: [],
      ercUpdates: [],
      rstkUpdates: [],
      uploadDialogVisible: false,
      uploadDialogVisible2: false,
      OkVisible: false,
      OkVisible2: false,
      stat: {},
      fromDates: null,
    };
  },
  created: function () {
    this.moment = moment;
  },
  methods: {
    saveRSTKToExcel() {
      if (!this.fromDates) {
        this.$notify.warning({
          title: "Выберите даты",
          message: "Для формирования отчёта необходимо указать период",
          offset: 150,
        });
        return;
      }
      console.log(this.tableData);
      UpdatesDataService.saveRSTKToExcel(this.fromDates)
        .then((response) => {
          let blob = new Blob([response.data], {
            type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          });
          let url = window.URL.createObjectURL(blob);
          let a = document.createElement("a");
          a.href = url;
          a.download = "Отчёт для ЕРЦ.xlsx";
          a.click();
          window.URL.revokeObjectURL(url);
        })
        .catch((e) => {
          console.log(e);
        });
    },
    syncERC() {
      this.uploadDialogVisible2 = true;
      this.OkVisible2 = false;
      UpdatesDataService.syncERC()
        .then((response) => {
          this.OkVisible2 = true;
          console.log(response);
          this.retrieveUpdates();
        })
        .catch((e) => {
          console.log(e);
        });
    },
    uploadStart(file) {
      this.uploadDialogVisible = true;
      this.OkVisible = false;
      console.log(file.filename);
    },
    uploadEnd(response, file) {
      this.OkVisible = true;
      console.log(response, file);
      this.retrieveUpdates();
    },
    retrieveUpdates() {
      UpdatesDataService.getAll()
        .then((response) => {
          this.ercUpdates = response.data["erc"];
          this.rstkUpdates = response.data["rstk"];
          this.stat = response.data["stat"];
          this.errorsData = response.data["errors_data"];
        })
        .catch((e) => {
          console.log(e);
        });
    },
    handleDelete(index, id) {
      console.log(index);
      UpdatesDataService.deleteReestrRstk(id)
        .then((response) => {
          console.log(response);
          this.retrieveUpdates();
        })
        .catch((e) => {
          console.log(e);
        });
      console.log("delete");
    },
  },
  mounted() {
    this.retrieveUpdates();
  },
};
</script>
