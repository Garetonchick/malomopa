#include <iostream>
#include <string>
#include <vector>
#include <thread>
#include <chrono>
#include <curl/curl.h>
#include <atomic>
#include <functional>

size_t write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    ((std::string*)userp)->append((char*)contents, size * nmemb);
    return size * nmemb;
}

void sendPostRequest(const std::string& url) {
    CURL* curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        std::string response_data;
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response_data);

        CURLcode res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            std::cerr << "CURL error: " << curl_easy_strerror(res) << std::endl;
        }
        curl_easy_cleanup(curl);
    }
}

int main() {
    const std::string baseUrl = "http://localhost:5252";
    const std::string baseUrl2 = "http://localhost:5253";
    auto assign_order_builder = [&baseUrl](int param) {
        return baseUrl + "/v1/assign_order?order-id=" + std::to_string(param) + "&executor-id=" + std::to_string(param);
    };
    auto acquire_order_builder = [&baseUrl2](int param) {
        return baseUrl2 + "/v1/acquire_order?executor-id=" + std::to_string(param);
    };
    auto cancel_order_builder = [&baseUrl](int param) {
        return baseUrl + "/v1/cancel_order?order-id=" + std::to_string(param);
    };
    auto assign_order_setup = []() {
    };
    auto acquire_order_setup = [&baseUrl, &assign_order_builder]() {
        for (int i = 0; i < 1'000; ++i) {
            sendPostRequest(assign_order_builder(i));
        }
    };
    auto cancel_order_setup = [&baseUrl, &assign_order_builder]() {
        for (int i = 0; i < 1'000; ++i) {
            sendPostRequest(assign_order_builder(i));
        }
    };
    const std::vector<std::string> name = {"assign_order", "acquire_order", "cancel_order"};
    const std::vector<std::function<void()>> setup = {assign_order_setup, acquire_order_setup, cancel_order_setup};
    const std::vector<std::function<std::string(int)>> builders = {assign_order_builder, acquire_order_builder, cancel_order_builder};
    const int durationSeconds = 10;
    const int threadsPerHandler = 12;

    for (size_t i = 0; i < 3; ++i) {
        std::cout << "Starting RPS test for handler " << name[i] << ". Setting up test." << std::endl;
        const auto& builder = builders[i];
        std::atomic<int> request_count = 0;
        setup[i]();
        std::cout << "Environment is ready." << std::endl;
        auto worker = [&]() {
            auto start = std::chrono::high_resolution_clock::now();
            while (true) {
                sendPostRequest(builder(request_count.fetch_add(1, std::memory_order::memory_order_acq_rel)));

                // Stop sending requests after durationSeconds
                if (std::chrono::high_resolution_clock::now() - start >= std::chrono::seconds(durationSeconds))
                    break;
            }
        };

        std::vector<std::thread> threads;
        for (int i = 0; i < threadsPerHandler; ++i) {
            threads.emplace_back(worker);
        }

        for (auto& t : threads) {
            t.join();
        }

        std::cout << "Handler: " << name[i]
                  << ", RPS: " << request_count.load(std::memory_order::memory_order_acquire) / durationSeconds << " requests/sec" << std::endl;
    }

    std::cout << "RPS test completed." << std::endl;
    return 0;
}
