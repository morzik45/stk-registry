import http from "../http-common";
import moment from "moment";

class ErcUpdatesDataService {
    getAll() {
        return http.get("/updates");
    }
    deleteReestrRstk(id) {
        return http.delete(`/updates/rstk/${id}`)
    }
    syncERC() {
        return http.post('/updates/uploadERC')
    }
    saveRSTKToExcel(dates) {
        dates = [moment(dates[0]).format("YYYY-MM-DD"), moment(dates[1]).format("YYYY-MM-DD")]
        return http.post(`/updates/make-rstk-excel`, JSON.stringify(dates), { responseType: 'arraybuffer' })
    }
}

export default new ErcUpdatesDataService();