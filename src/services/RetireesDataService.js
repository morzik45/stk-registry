import http from "../http-common";

class RetireesDataService {
    getAll() {
        return http.get("/retiree");
    }
    find(searchStr) {
        return http.get(`/retiree?search=${searchStr}`);
    }
}

export default new RetireesDataService();